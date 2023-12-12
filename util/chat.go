package util

import (
	"github.com/dlclark/regexp2"
	"github.com/sashabaranov/go-openai"
	"strings"
)

func GetOpenAIChatMessages(chatContext string) []openai.ChatCompletionMessage {
	ctx := chatContext + "\n\n[system](#sydney__placeholder)"
	re := regexp2.MustCompile(`\[(system|user|assistant)]\(#(.*?)\)([\s\S]*?)(?=\n.*?(^\[(system|user|assistant)]\(#.*?\)))`,
		regexp2.IgnoreCase|regexp2.Multiline)
	var result []openai.ChatCompletionMessage
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
		if groups[2].String() != "message" && groups[2].String() != "additional_instructions" {
			content = groups[2].String() + "\n" + content
		}
		result = append(result, openai.ChatCompletionMessage{
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
