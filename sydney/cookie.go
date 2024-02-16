package sydney

import (
	"errors"
	"regexp"
	"strconv"
	"sydneyqt/util"
	"time"
)

func (o *Sydney) GetUser() (string, error) {
	_, client, err := util.MakeHTTPClient(o.proxy, 15*time.Second)
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
	resp, err := client.R().
		SetHeader("Cookie", util.FormatCookieString(cookies)).
		Get("https://www.bing.com/search?q=Bing+AI&showconv=1")
	if err != nil {
		return "", err
	}
	if resp.GetStatusCode() != 200 {
		return "", errors.New("http status code is not 200: " + strconv.Itoa(resp.GetStatusCode()))
	}
	respText := resp.String()
	arr := regexp.MustCompile(`data-clarity-mask="true" title="(.*?)"`).FindStringSubmatch(respText)
	if len(arr) < 2 {
		return "", errors.New("cannot identify current user, please check if cookie is expired")
	}
	return arr[1], nil
}
