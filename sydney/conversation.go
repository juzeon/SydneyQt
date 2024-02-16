package sydney

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"sydneyqt/util"
	"time"
)

func (o *Sydney) createConversation() (CreateConversationResponse, error) {
	var empty CreateConversationResponse
	_, client, err := util.MakeHTTPClient(o.proxy, 10*time.Second)
	if err != nil {
		return empty, err
	}
	resp, err := client.R().SetHeader("Accept", "application/json").
		SetHeader("Cookie", util.FormatCookieString(o.cookies)).Get(o.createConversationURL)
	if err != nil {
		return empty, err
	}
	bodyV := resp.Bytes()
	if resp.GetStatusCode() != 200 {
		slog.Error("Failed body", "v", string(bodyV))
		return empty, errors.New("failed to create the conversation, code: " +
			strconv.Itoa(resp.StatusCode) + "; please check your proxy settings and your account")
	}
	var response CreateConversationResponse
	err = json.Unmarshal(bodyV, &response)
	if err != nil {
		return empty, err
	}
	if response.Result.Value != "Success" {
		return empty, errors.New("failed to create the conversation: message: " + response.Result.Message)
	}
	if value := resp.Header.Get("X-Sydney-Encryptedconversationsignature"); value != "" {
		response.SecAccessToken = value
	}
	var cookieFields []string
	for _, field := range resp.Header.Values("set-cookie") {
		cookieFields = append(cookieFields, strings.Split(field, ";")[0])
	}
	newCookies := util.ParseCookiesFromString(strings.Join(cookieFields, "; "))
	slog.Info("Cookies to update when creating conversation", "diff", newCookies)
	o.UpdateModifiedCookies(newCookies)
	slog.Debug("Create conversation", "response", response)
	slog.Info("Created Conversation")
	return response, nil
}
