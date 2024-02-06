package sydney

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"log/slog"
	"sydneyqt/sydney/internal/hex"
	"sydneyqt/util"
	"time"
)

func (o *Sydney) BypassCaptcha(
	stopCtx context.Context, conversationID string, messageID string,
) (cookies map[string]string, err error) {
	if o.bypassServer == "" {
		return nil, errors.New("no bypass server specified")
	}
	hClient, err := util.MakeHTTPClient(o.proxy, 0)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	var response BypassCaptchaResponse
	err = json.Unmarshal(resp.Body(), &response)
	if err != nil {
		return nil, err
	}
	slog.Debug("Bypass captcha response", "v", response)
	if response.Error != "" {
		return nil, errors.New("bypass captcha error: " + response.Error)
	}
	cookies = util.ParseCookiesFromString(response.Result.Cookies)
	// new cookies: cct, GC, _C_ETH=1, _C_Auth=
	return cookies, nil
}
