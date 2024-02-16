package sydney

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"sydneyqt/util"
	"time"
)

func (o *Sydney) createConversation() (CreateConversationResponse, error) {
	client, err := util.MakeHTTPClient(o.proxy, 10*time.Second)
	emptyResponse := CreateConversationResponse{}
	if err != nil {
		return emptyResponse, err
	}
	req, err := http.NewRequest("GET", o.createConversationURL, nil)
	if err != nil {
		return emptyResponse, err
	}
	for k, v := range o.headersCreateConversation() {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return emptyResponse, err
	}
	defer resp.Body.Close()
	bodyV, err := io.ReadAll(resp.Body)
	if err != nil {
		return emptyResponse, err
	}
	if resp.StatusCode != 200 {
		slog.Error("Failed body", "v", string(bodyV))
		return emptyResponse, errors.New("failed to create the conversation, code: " +
			strconv.Itoa(resp.StatusCode) + "; please check your proxy settings and your account")
	}
	var response CreateConversationResponse
	err = json.Unmarshal(bodyV, &response)
	if err != nil {
		return emptyResponse, err
	}
	if response.Result.Value != "Success" {
		return emptyResponse, errors.New("failed to create the conversation: message: " + response.Result.Message)
	}
	if value := resp.Header.Get("X-Sydney-Encryptedconversationsignature"); value != "" {
		response.SecAccessToken = value
	}
	slog.Debug("Create conversation", "response", response)
	slog.Info("Created Conversation")
	return response, nil
}
