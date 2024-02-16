package sydney

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	"github.com/google/uuid"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sydneyqt/sydney/internal/hex"
	"sydneyqt/util"
	"time"
)

func (o *Sydney) ResolveCaptcha(stopCtx context.Context) (err error) {
	defer func() {
		if err0 := recover(); err0 != nil {
			slog.Warn("Error resolving captcha", "err", err0)
			err = err0.(error)
		}
	}()
	iframeID := uuid.New().String()
	l := launcher.NewUserMode().Context(stopCtx).
		Leakless(true).
		UserDataDir(filepath.Join(os.TempDir(), "rod-user-data-"+uuid.New().String())).
		Set("disable-default-apps").
		Set("no-first-run").Headless(false)
	defer l.Cleanup()
	u := l.MustLaunch()
	browser := rod.New().Context(stopCtx).NoDefaultDevice().ControlURL(u).MustConnect()
	defer browser.MustClose()
	var cookies []*proto.NetworkCookie
	for k, v := range o.cookies {
		cookies = append(cookies, &proto.NetworkCookie{
			Name:    k,
			Value:   v,
			Domain:  ".bing.com",
			Path:    "/",
			Expires: proto.TimeSinceEpoch(time.Now().Add(1 * time.Hour).Unix()),
		})
	}
	browser.MustSetCookies(cookies...)
	page := stealth.MustPage(browser)
	page.MustNavigate("https://www.bing.com/turing/captcha/challenge?" +
		"q=&iframeid=local-gen-" + iframeID)
	page.MustElement("body")
	page.MustEval("()=>{let info=document.createElement('h3');" +
		"info.textContent='↑ Please help click if this cannot be processed automatically!';" +
		"document.body.appendChild(info);}")
	router := page.HijackRequests()
	waitCh := make(chan struct{}, 16)
	defer close(waitCh)
	var resCookies map[string]string
	router.MustAdd("https://www.bing.com/challenge/verify*", func(hijack *rod.Hijack) {
		hijack.MustLoadResponse()
		for key, values := range hijack.Response.Headers() {
			if strings.ToLower(key) != "set-cookie" {
				continue
			}
			var arr []string
			for _, v := range values {
				arr = append(arr, strings.Split(v, ";")[0])
			}
			resCookies = util.ParseCookiesFromString(strings.Join(arr, "; "))
		}
		waitCh <- struct{}{}
	})
	go router.Run()
	defer router.Stop()
	select {
	case <-time.Tick(60 * time.Second):
		return errors.New("timeout verifying challenge token")
	case <-stopCtx.Done():
		return stopCtx.Err()
	case <-waitCh:
	}
	slog.Info("Captcha resCookies", "v", resCookies)
	if err := o.postprocessCaptchaCookies(resCookies); err != nil {
		return err
	}
	return nil
}
func (o *Sydney) BypassCaptcha(stopCtx context.Context, conversationID string, messageID string) error {
	if o.bypassServer == "" {
		return errors.New("no bypass server specified")
	}
	hClient, err := util.MakeHTTPClient(o.proxy, 0)
	if err != nil {
		return err
	}
	client := resty.New().SetTransport(hClient.Transport).SetTimeout(60 * time.Second)
	req := BypassCaptchaRequest{
		IG:       hex.NewUpperHex(32),
		Cookies:  util.FormatCookieString(o.cookies),
		IFrameID: "local-gen-" + uuid.New().String(),
		ConvID:   conversationID,
		RID:      messageID,
	}
	slog.Debug("Bypass CAPTCHA request", "v", req)
	resp, err := client.R().SetContext(stopCtx).SetBody(req).Post(o.bypassServer)
	if err != nil {
		return fmt.Errorf("cannot communicate with captcha bypass server: %w", err)
	}
	slog.Debug("Bypass captcha response body", "v", string(resp.Body()))
	var response BypassCaptchaResponse
	err = json.Unmarshal(resp.Body(), &response)
	if err != nil {
		return fmt.Errorf("cannot unmarshal json from captcha bypass server: %w", err)
	}
	if response.Error != "" {
		return errors.New("bypass captcha error: " + response.Error)
	}
	cookies := util.ParseCookiesFromString(response.Result.Cookies)
	if err := o.postprocessCaptchaCookies(cookies); err != nil {
		return fmt.Errorf("%w; screenshot: "+
			strings.TrimSuffix(o.bypassServer, "/")+
			response.Result.ScreenShot, err)
	}
	return nil
}
func (o *Sydney) UpdateModifiedCookies(modifiedCookies map[string]string) {
	for k, v := range modifiedCookies { // keep the map pointer
		o.cookies[k] = v
	}
	err := util.UpdateCookiesFile(o.cookies)
	if err != nil {
		slog.Warn("Cannot update cookies file: ", "err", err)
	}
}
func (o *Sydney) postprocessCaptchaCookies(modifiedCookies map[string]string) error {
	if _, ok := modifiedCookies["cct"]; !ok {
		return errors.New("captcha cookies not valid: no cookie named cct found")
	}
	o.UpdateModifiedCookies(modifiedCookies)
	return nil
}
