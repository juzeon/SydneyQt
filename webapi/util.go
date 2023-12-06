package main

import (
	"strings"
)

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
