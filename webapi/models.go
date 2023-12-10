package main

import "sydneyqt/sydney"

type CreateConversationRequest struct {
	Cookies string `json:"cookies,omitempty"`
}

type ChatStreamRequest struct {
	Prompt            string                            `json:"prompt"`
	WebpageContext    string                            `json:"context"`
	Conversation      sydney.CreateConversationResponse `json:"conversation,omitempty"`
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

type ChoiceDelta struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionChunkChoice struct {
	Index        int         `json:"index"`
	Delta        ChoiceDelta `json:"delta"`
	FinishReason *string      `json:"finish_reason"`
}

type OpenAIChatCompletionChunk struct {
	ID                string                      `json:"id"`
	Object            string                      `json:"object"`
	Created           int64                       `json:"created"`
	Model             string                      `json:"model"`
	SystemFingerprint string                      `json:"system_fingerprint"`
	Choices           []ChatCompletionChunkChoice `json:"choices"`
}

type ChoiceMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type UsageStats struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatCompletionChoice struct {
	Index        int           `json:"index"`
	Message      ChoiceMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

type OpenAIChatCompletion struct {
	ID                string                 `json:"id"`
	Object            string                 `json:"object"`
	Created           int64                  `json:"created"`
	Model             string                 `json:"model"`
	SystemFingerprint string                 `json:"system_fingerprint"`
	Choices           []ChatCompletionChoice `json:"choices"`
	Usage             UsageStats             `json:"usage"`
}
