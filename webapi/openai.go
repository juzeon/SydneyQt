package main

import (
	"errors"
	"strings"
	"sydneyqt/sydney"
	"time"
)

var (
	ErrMissingPrompt     = errors.New("user prompt is missing (last message is not sent by user)")
	FinishReasonStop     = "stop"
	FinishReasonLength   = "length"
	MessageRoleUser      = "user"
	MessageRoleAssistant = "assistant"
	MessageRoleSystem    = "system"
)

func ParseOpenAIMessages(messages []OpenAIMessage) (OpenAIMessagesParseResult, error) {
	if len(messages) == 0 {
		return OpenAIMessagesParseResult{}, ErrMissingPrompt
	}

	// find the last user message
	var promptIndex int
	var promptMessage OpenAIMessage

	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == MessageRoleUser {
			promptIndex = i
			promptMessage = messages[i]
			break
		}
	}

	prompt, imageUrl := ParseOpenAIMessageContent(promptMessage.Content)

	if prompt == "" {
		return OpenAIMessagesParseResult{}, ErrMissingPrompt
	}

	if len(messages) == 1 {
		return OpenAIMessagesParseResult{
			Prompt:   prompt,
			ImageURL: imageUrl,
		}, nil
	}

	// exclude the promptMessage from the array
	messages = append(messages[:promptIndex], messages[promptIndex+1:]...)

	// construct context
	var contextBuilder strings.Builder
	contextBuilder.WriteString("\n\n")

	for i, message := range messages {
		// assert types
		text, _ := ParseOpenAIMessageContent(message.Content)

		// append role to context
		switch message.Role {
		case MessageRoleUser:
			contextBuilder.WriteString("[user](#message)\n")
		case MessageRoleAssistant:
			contextBuilder.WriteString("[assistant](#message)\n")
		case MessageRoleSystem:
			contextBuilder.WriteString("[system](#additional_instructions)\n")
		default:
			continue // skip unknown roles
		}

		// append content to context
		contextBuilder.WriteString(text)
		if i != len(messages)-1 {
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
	case []interface{}:
		// content is array of objects, and it contains prompt and optional image url
		for _, content := range content {
			content, ok := content.(map[string]interface{})
			if !ok {
				continue
			}
			switch content["type"] {
			case "text":
				if contentText, ok := content["text"].(string); ok {
					text = contentText
				}
			case "image_url":
				if url, ok := content["image_url"].(map[string]interface{}); ok {
					imageUrl, _ = url["url"].(string)
				}
			}
		}
	}

	return
}

func NewOpenAIChatCompletion(model, content, finishReason string) *OpenAIChatCompletion {
	return &OpenAIChatCompletion{
		ID:                "chatcmpl-123",
		Object:            "chat.completion",
		Created:           time.Now().Unix(),
		Model:             model,
		SystemFingerprint: "fp_44709d6fcb",
		Choices: []ChatCompletionChoice{
			{
				Index: 0,
				Message: ChoiceMessage{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: finishReason,
			},
		},
		Usage: UsageStats{
			PromptTokens:     1024,
			CompletionTokens: 1024,
			TotalTokens:      2048,
		},
	}
}

func NewOpenAIChatCompletionChunk(model, delta string, finishReason *string) *OpenAIChatCompletionChunk {
	return &OpenAIChatCompletionChunk{
		ID:                "chatcmpl-123",
		Object:            "chat.completion",
		Created:           time.Now().Unix(),
		Model:             model,
		SystemFingerprint: "fp_44709d6fcb",
		Choices: []ChatCompletionChunkChoice{
			{
				Index: 0,
				Delta: ChoiceDelta{
					Role:    "assistant",
					Content: delta,
				},
				FinishReason: finishReason,
			},
		},
	}
}

func ToOpenAIImageGeneration(result sydney.GenerateImageResult) OpenAIImageGeneration {
	var objects []OpenAIImageObject

	for _, url := range result.ImageURLs {
		urlWithoutQuery := strings.Split(url, "?")[0]
		objects = append(objects, OpenAIImageObject{
			URL:           urlWithoutQuery,
			RevisedPrompt: result.Text,
		})
	}

	return OpenAIImageGeneration{
		Created: time.Now().Unix(),
		Data:    objects,
	}
}
