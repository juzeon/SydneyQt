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
		Cookies:  o.formattedCookies,
		IFrameID: "local-gen-" + uuid.New().String(),
		ConvID:   conversationID,
		RID:      messageID,
	}
	slog.Info("Bypass CAPTCHA request", "v", req)
	resp, err := client.R().SetContext(stopCtx).SetBody(req).Post(o.bypassServer)
	if err != nil {
		return err
	}
	var response BypassCaptchaResponse
	err = json.Unmarshal(resp.Body(), &response)
	if err != nil {
		return err
	}
	slog.Info("Bypass captcha response", "v", response)
	if response.Error != "" {
		return errors.New("bypass captcha error: " + response.Error)
	}
	return nil
}
