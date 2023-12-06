package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOpenAIMessages(t *testing.T) {
	t.Run("valid with image", func(t *testing.T) {
		messages := []OpenAIMessage{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "Hello!",
					},
				},
			},
			{
				"role": "assistant",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "Hi!",
					},
				},
			},
			{
				"role": "user",
				"content": []map[string]interface{}{
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
			WebpageContext: "[user](#message)\nHello!\n\n[assistant](#message)\nHi!",
			ImageURL:       "https://example.com/image.jpg",
		}, result)
	})
}
