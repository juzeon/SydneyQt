package sydney

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"log/slog"
	"net/url"
	"regexp"
	"sydneyqt/util"
	"time"
)

func (o *Sydney) GenerateImage(generativeImage GenerativeImage) (GenerateImageResult, error) {
	start := time.Now()
	var empty GenerateImageResult
	hClient, err := util.MakeHTTPClient(o.proxy, 0)
	if err != nil {
		return empty, err
	}
	client := resty.New().SetTransport(hClient.Transport).
		SetHeaders(o.headersCreateImage).
		SetTimeout(15 * time.Second)
	resp, err := client.R().Get(generativeImage.URL)
	if err != nil && !errors.Is(err, resty.ErrAutoRedirectDisabled) {
		return empty, err
	}
	arr := regexp.MustCompile("/images/create/async/results/(.*?)\\?").FindStringSubmatch(string(resp.Body()))
	if len(arr) < 2 {
		return empty, errors.New("cannot find image creation result")
	}
	resultID := arr[1]
	re := regexp.MustCompile(`<img class="mimg".*?src="(.*?)"`)
	u := "https://www.bing.com/images/create/async/results/" + resultID +
		"?q=" + url.QueryEscape(generativeImage.Text) + "&partner=sydney&showselective=1&IID=images.as"
	slog.Info("Result URL", "v", u)
	for i := 0; i < 15; i++ {
		time.Sleep(3 * time.Second)
		resp, err := client.R().Get(u)
		if err != nil {
			return empty, err
		}
		var imageURLs []string
		arr := re.FindAllStringSubmatch(string(resp.Body()), -1)
		if len(arr) == 0 {
			slog.Info("No matched images currently", "body", string(resp.Body()))
			continue
		}
		for _, match := range arr {
			imageURLs = append(imageURLs, match[1])
		}
		return GenerateImageResult{
			GenerativeImage: generativeImage,
			ImageURLs:       imageURLs,
			Duration:        time.Now().Sub(start),
		}, nil
	}
	return empty, errors.New("image creation timeout")
}
