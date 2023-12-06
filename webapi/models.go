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
