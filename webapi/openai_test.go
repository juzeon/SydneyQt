package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOpenAIMessages(t *testing.T) {
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
