package sydney

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"sydneyqt/util"
	"time"
)

func (o *Sydney) CreateConversation() (CreateConversationResponse, error) {
	client, err := util.MakeHTTPClient(o.proxy, 10*time.Second)
	if err != nil {
		return CreateConversationResponse{}, err
	}
	req, err := http.NewRequest("GET", o.createConversationURL, nil)
	if err != nil {
		return CreateConversationResponse{}, err
	}
	for k, v := range o.headersCreateConversation {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return CreateConversationResponse{}, err
	}
	defer resp.Body.Close()
	bodyV, err := io.ReadAll(resp.Body)
	if err != nil {
		return CreateConversationResponse{}, err
	}
	if resp.StatusCode != 200 {
		return CreateConversationResponse{}, errors.New("Authentication failed: " + string(bodyV))
	}
	var response CreateConversationResponse
	err = json.Unmarshal(bodyV, &response)
	if err != nil {
		return CreateConversationResponse{}, err
	}
	if response.Result.Value == "UnauthorizedRequest" {
		return CreateConversationResponse{}, errors.New(response.Result.Message)
	}
	if value := resp.Header.Get("X-Sydney-Encryptedconversationsignature"); value != "" {
		response.SecAccessToken = value
	}
	if o.debug {
		slog.Info("Create conversation", "response", response)
	}
	slog.Info("Created Conversation")
	return response, nil
}
