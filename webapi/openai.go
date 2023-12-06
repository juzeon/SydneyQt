package main

import (
	"errors"
	"strings"
)

var (
	ErrInvalidOpenAIMessage = errors.New("invalid openai message")
	ErrUnknownPrompt        = errors.New("unknown prompt")
)

func ParseOpenAIMessages(messages []OpenAIMessage) (OpenAIMessagesParseResult, error) {
	if len(messages) == 0 {
		return OpenAIMessagesParseResult{}, nil
	}

	// last message must be user prompt
	lastMessage := messages[len(messages)-1]
	prompt, imageUrl := ParseOpenAIMessageContent(lastMessage.Content)

	if lastMessage.Role != "user" || prompt == "" {
		return OpenAIMessagesParseResult{}, ErrUnknownPrompt
	}

	// construct context
	var contextBuilder strings.Builder
	contextBuilder.WriteString("\n\n")

	for i, message := range messages[:len(messages)-1] {
		// assert types
		text, _ := ParseOpenAIMessageContent(message.Content)

		// append role to context
		switch message.Role {
		case "user":
			contextBuilder.WriteString("[user](#message)\n")
		case "assistant":
			contextBuilder.WriteString("[assistant](#message)\n")
		case "system":
			contextBuilder.WriteString("[system](#additional_instructions)\n")
		default:
			continue // skip unknown roles
		}

		// append content to context
		contextBuilder.WriteString(text)
		if i != len(messages)-2 {
			contextBuilder.WriteString("\n\n")
		}
	}

	return OpenAIMessagesParseResult{
		Prompt:         prompt,
		WebpageContext: contextBuilder.String(),
		ImageURL:       imageUrl,
	}, nil
}

func ParseOpenAIMessageContent(content interface{}) (text, imageUrl string) {
	switch content := content.(type) {
	case string:
		// content is string, and it automatically becomes prompt
		text = content
	case []map[string]interface{}:
		// content is array of objects, and it contains prompt and optional image url
		for _, content := range content {
			switch content["type"] {
			case "text":
				if contentText, ok := content["text"].(string); ok {
					text = contentText
				}
			case "image_url":
				if url, ok := content["image_url"].(map[string]string); ok {
					imageUrl = url["url"]
				}
			}
		}
	}

	return
}
