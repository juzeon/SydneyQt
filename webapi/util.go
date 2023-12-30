package main

import (
	"strings"
	"encoding/json"
	"os"
	"errors"
)

type FileCookie struct {
	Domain         string      `json:"domain"`
	ExpirationDate float64     `json:"expirationDate"`
	HostOnly       bool        `json:"hostOnly"`
	HttpOnly       bool        `json:"httpOnly"`
	Name           string      `json:"name"`
	Path           string      `json:"path"`
	SameSite       string      `json:"sameSite"`
	Secure         bool        `json:"secure"`
	Session        bool        `json:"session"`
	StoreId        interface{} `json:"storeId"`
	Value          string      `json:"value"`
}

func ReadCookiesFile() (map[string]string, error) {
	res := map[string]string{}
	v, err := os.ReadFile("cookies.json")
	if err != nil {
		return res, err
	}
	var cookies []FileCookie
	err = json.Unmarshal(v, &cookies)
	if err != nil {
		return res, errors.New("failed to json.Unmarshal content of cookie file")
	}
	for _, cookie := range cookies {
		res[cookie.Name] = cookie.Value
	}
	return res, nil
}

func ParseCookies(cookiesStr string) map[string]string {
	cookies := map[string]string{}
	for _, cookie := range strings.Split(cookiesStr, ";") {
		parts := strings.Split(cookie, "=")
		if len(parts) == 2 {
			cookies[parts[0]] = parts[1]
		}
	}
	
	return cookies
}
