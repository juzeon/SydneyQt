package main

import (
	"strings"
	"io/ioutil"
	"fmt"
	"encoding/json"
)

func ParseCookies(cookiesStr string) map[string]string {
	cookies := map[string]string{}
	
	if strings.HasSuffix(cookiesStr, ".json") {
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
