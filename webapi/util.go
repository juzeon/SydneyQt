package main

import (
	"strings"
	"io/ioutil"
	"fmt"
	"encoding/json"
	"os"
)

func ParseCookies(cookiesStr string) map[string]string {
	cookies := map[string]string{}
	
	// use the os.Stat function to make sure the file exists and make sure it points to a .json file
	// else load it is a cookieString from a header
	if _, err := os.Stat(cookiesStr); err == nil && strings.HasSuffix(cookiesStr, ".json") {
		data, err := ioutil.ReadFile(cookiesStr)
		if err != nil {
			fmt.Println("error:", err)
		}

		type Item struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		}

		var items []Item
		err = json.Unmarshal(data, &items)
		if err != nil {
			fmt.Println(err)
		}
		
		for _, item := range items {
			cookies[item.Name] = item.Value
		}
	} else {
		for _, cookie := range strings.Split(cookiesStr, ";") {
			parts := strings.Split(cookie, "=")
			if len(parts) == 2 {
			cookies[parts[0]] = parts[1]
			}
		}
	}

	
	return cookies
}
