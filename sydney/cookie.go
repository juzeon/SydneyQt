package sydney

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"regexp"
	"strconv"
	"sydneyqt/util"
	"time"
)

func (o *Sydney) GetUser() (string, error) {
	hClient, err := util.MakeHTTPClient(o.proxy, 0)
	if err != nil {
		return "", err
	}
	cookies, err := util.ReadCookiesFile()
	if err != nil {
		return "", err
	}
	if len(cookies) == 0 {
		return "", errors.New("cookie file is empty")
	}
	client := resty.New().SetTransport(hClient.Transport).SetTimeout(15 * time.Second)
	resp, err := client.R().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 Edg/114.0.1788.0").
		SetHeader("Cookie", util.FormatCookieString(cookies)).
		Get("https://www.bing.com/search?q=Bing+AI&showconv=1")
	if err != nil {
		return "", err
	}
	if resp.StatusCode() != 200 {
		return "", errors.New("http status code is not 200: " + strconv.Itoa(resp.StatusCode()))
	}
	respText := string(resp.Body())
	arr := regexp.MustCompile(`data-clarity-mask="true" title="(.*?)"`).FindStringSubmatch(respText)
	if len(arr) < 2 {
		return "", errors.New("cannot identify current user, please check if cookie is expired")
	}
	return arr[1], nil
}
