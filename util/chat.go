package util

import (
	"github.com/dlclark/regexp2"
	"github.com/sashabaranov/go-openai"
	"strings"
)

func GetOpenAIChatMessages(chatContext string) []openai.ChatCompletionMessage {
	var result []openai.ChatCompletionMessage
	messages := GetChatMessage(chatContext)
	for _, msg := range messages {
		content := msg.Content
		if msg.Type != "message" && msg.Type != "additional_instructions" {
			content = "# " + msg.Type + "\n" + content
		}
		result = append(result, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: content,
		})
	}
	return result
}

type ChatMessage struct {
	Role    string `json:"role"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

func GetChatMessage(chatContext string) []ChatMessage {
	ctx := chatContext + "\n\n[system](#sydney__placeholder)"
	re := regexp2.MustCompile(`\[(system|user|assistant)]\(#(.*?)\)([\s\S]*?)(?=\n.*?(^\[(system|user|assistant)]\(#.*?\)))`,
		regexp2.IgnoreCase|regexp2.Multiline)
	var result []ChatMessage
	match, err := re.FindStringMatch(ctx)
	if err != nil {
		GracefulPanic(err)
	}
	for match != nil {
		groups := match.Groups()
		if groups[2].String() == "sydney__placeholder" {
			continue
		}
		content := strings.TrimSpace(groups[3].String())
		result = append(result, ChatMessage{
			Type:    groups[2].String(),
			Role:    groups[1].String(),
			Content: content,
		})
		match, err = re.FindNextMatch(match)
		if err != nil {
			GracefulPanic(err)
		}
	}
	return result
}
func CreateOpenAIClient(proxy string, key string, endpoint string) (*openai.Client, error) {
	hClient, err := MakeHTTPClient(proxy, 0)
	if err != nil {
		return nil, err
	}
	config := openai.DefaultConfig(key)
	config.BaseURL = endpoint
	config.HTTPClient = hClient
	return openai.NewClientWithConfig(config), nil
}
