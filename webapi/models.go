package main

import "sydneyqt/sydney"

type CreateConversationRequest struct {
	Cookies string `json:"cookies,omitempty"`
}

type ChatStreamRequest struct {
	Conversation      sydney.CreateConversationResponse `json:"conversation"`
	Prompt            string                            `json:"prompt"`
	WebpageContext    string                            `json:"context"`
	Cookies           string                            `json:"cookies,omitempty"`
	ImageURL          string                            `json:"imageUrl,omitempty"`
	NoSearch          bool                              `json:"noSearch,omitempty"`
	ConversationStyle string                            `json:"conversationStyle,omitempty"`
	Locale            string                            `json:"locale,omitempty"`
}

// The `content` field can have different types
// Example:
//
//	{
//		"role": "user",
//		"content": "Hello!"
//	}
//
// or
//
//	{
//		"role": "user",
//		"content": [
//			{
//				"type": "text",
//				"text": "What’s in this image?"
//			},
//			{
//				"type": "image_url",
//				"image_url": {
//					"url": "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"
//				}
//			}
//		]
//	}
type OpenAIMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type OpenAIMessagesParseResult struct {
	WebpageContext string
	Prompt         string
	ImageURL       string
}

// Most fields are omitted due to limitations of the Bing API
type OpenAIChatCompletionRequest struct {
	Model        string                            `json:"model"`
	Messages     []OpenAIMessage                   `json:"messages"`
	Stream       bool                              `json:"stream"`
	Conversation sydney.CreateConversationResponse `json:"conversation,omitempty"`
}

type OpenAIChatCompletionChunk []struct {
	ID                string `json:"id"`
	Object            string `json:"object"`
	Created           int    `json:"created"`
	Model             string `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
	Choices           []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

type OpenAIChatCompletion []struct {
	ID                string `json:"id"`
	Object            string `json:"object"`
	Created           int    `json:"created"`
	Model             string `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
	Choices           []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}
