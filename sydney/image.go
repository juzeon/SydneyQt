package sydney

import (
	"errors"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"sydneyqt/util"
	"time"
)

func (o *Sydney) GenerateImage(generativeImage GenerativeImage) (GenerateImageResult, error) {
	start := time.Now()
	var empty GenerateImageResult
	_, client, err := util.MakeHTTPClient(o.proxy, 15*time.Second)
	if err != nil {
		return empty, err
	}
	client.SetCommonHeader("Cookie", util.FormatCookieString(o.cookies))
	resp, err := client.R().Get(generativeImage.URL)
	if err != nil {
		return empty, err
	}
	arr := regexp.MustCompile("/images/create/async/results/(.*?)\\?").FindStringSubmatch(resp.String())
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
		bodyStr := resp.String()
		if strings.Contains(bodyStr, "Please try again or come back later") {
			return empty, errors.New("the prompt for image creation has been rejected by Bing")
		}
		var imageURLs []string
		arr := re.FindAllStringSubmatch(bodyStr, -1)
		if len(arr) == 0 {
			slog.Info("No matched images currently", "body", bodyStr)
			continue
		}
		for _, match := range arr {
			imageURLs = append(imageURLs, match[1])
		}
		slog.Info("Created images successfully", "images", imageURLs)
		return GenerateImageResult{
			GenerativeImage: generativeImage,
			ImageURLs:       imageURLs,
			Duration:        time.Now().Sub(start),
		}, nil
	}
	return empty, errors.New("image creation timeout")
}
