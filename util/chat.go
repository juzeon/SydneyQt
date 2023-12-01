package util

import (
	"github.com/sashabaranov/go-openai"
	"regexp"
	"strings"
)

func GetOpenAIChatMessages(chatContext string) []openai.ChatCompletionMessage {
	ctx := chatContext + "\n\n[system](#sydney__placeholder)"
	re := regexp.MustCompile(`(?m)\[(system|user|assistant)]\(#(.*?)\)([\s\S]*?)(?=\n.*?(^\[(system|user|assistant)]\(#.*?\)))`)
	matches := re.FindAllStringSubmatch(ctx, -1)
	var result []openai.ChatCompletionMessage
	for _, v := range matches {
		if v[2] != "sydney__placeholder" {
			content := strings.TrimSpace(v[3])
			if v[2] != "message" && v[2] != "additional_instructions" {
				content = v[2] + "\n" + content
			}
			result = append(result, openai.ChatCompletionMessage{
				Role:    v[1],
				Content: content,
			})
		}
	}
	return result
}
