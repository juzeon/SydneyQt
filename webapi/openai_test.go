package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOpenAIMessages(t *testing.T) {
	t.Run("parse json", func(t *testing.T) {
		jsonString := "{\n  \"model\": \"gpt-3.5-turbo\",\n  \"messages\": [\n    {\n      \"role\": \"user\",\n      \"content\": [\n        {\n          \"type\": \"text\",\n          \"text\": \"What's in the image?\"\n        },\n        {\n          \"type\": \"image_url\",\n          \"image_url\": {\n            \"url\": \"https://th.bing.com/th/id/OIP.S2vzW4WYmmR0Fq54jCJ5WAAAAA?&rs=1&pid=ImgDetMain\"\n          }\n        }\n      ]\n    }\n  ],\n  \"stream\": true\n}"
		var request OpenAIChatCompletionRequest
		err := json.Unmarshal([]byte(jsonString), &request)
		assert.Nil(t, err)

		result, err := ParseOpenAIMessages(request.Messages)
		assert.Nil(t, err)
		assert.Equal(t, OpenAIMessagesParseResult{
			Prompt:         "What's in the image?",
			WebpageContext: "",
			ImageURL:       "https://th.bing.com/th/id/OIP.S2vzW4WYmmR0Fq54jCJ5WAAAAA?&rs=1&pid=ImgDetMain",
		}, result)
	})
	t.Run("empty", func(t *testing.T) {
		result, err := ParseOpenAIMessages([]OpenAIMessage{})
		assert.Equal(t, ErrMissingPrompt, err)
		assert.Equal(t, OpenAIMessagesParseResult{}, result)
	})
	t.Run("single message", func(t *testing.T) {
		messages := []OpenAIMessage{
			{
				Role:    "user",
				Content: "Hello!",
			},
		}
		result, err := ParseOpenAIMessages(messages)
		assert.Nil(t, err)
		assert.Equal(t, OpenAIMessagesParseResult{
			Prompt: "Hello!",
		}, result)
	})
	t.Run("not ending with user message", func(t *testing.T) {
		messages := []OpenAIMessage{
			{
				Role:    "user",
				Content: "Hello!",
			},
			{
				Role:    "assistant",
				Content: "Hi!",
			},
		}
		result, err := ParseOpenAIMessages(messages)
		assert.Nil(t, err)
		assert.Equal(t, "Hello!", result.Prompt)
	})
	t.Run("valid with image", func(t *testing.T) {
		messages := []OpenAIMessage{
			{
				Role:    "system",
				Content: "You are Sydney.",
			},
			{
				Role:    "user",
				Content: "Hello!",
			},
			{
				Role:    "assistant",
				Content: "Hi!",
			},
			{
				Role: "user",
				Content: []map[string]interface{}{
					{
						"type": "text",
						"text": "How are you?",
					},
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": "https://example.com/image.jpg",
						},
					},
				},
			},
		}
		result, err := ParseOpenAIMessages(messages)
		assert.Nil(t, err)
		assert.Equal(t, OpenAIMessagesParseResult{
			Prompt:         "How are you?",
			WebpageContext: "\n\n[system](#additional_instructions)\nYou are Sydney.\n\n[user](#message)\nHello!\n\n[assistant](#message)\nHi!",
			ImageURL:       "https://example.com/image.jpg",
		}, result)
	})
}
