package sydney

import (
	"context"
	"errors"
)

const delimiter = '\x1e'

type LocationHint struct {
	Country           string `json:"country"`
	State             string `json:"state"`
	City              string `json:"city"`
	TimezoneOffset    int    `json:"timezoneoffset"`
	CountryConfidence int    `json:"countryConfidence"`
	Center            LatLng `json:"Center"`
	RegionType        int    `json:"RegionType"`
	SourceType        int    `json:"SourceType"`
}
type LatLng struct {
	Latitude  float64 `json:"Latitude"`
	Longitude float64 `json:"Longitude"`
}
type CreateConversationResult struct {
	Value   string `json:"value"`
	Message string `json:"message"`
}
type CreateConversationResponse struct {
	ConversationId        string                   `json:"conversationId"`
	ClientId              string                   `json:"clientId"`
	Result                CreateConversationResult `json:"result"`
	SecAccessToken        string                   `json:"secAccessToken"`
	ConversationSignature string                   `json:"conversationSignature"`
}
type RawMessage struct {
	Data  string
	Error error
}

const (
	MessageTypeSearchQuery        = "search_query"
	MessageTypeSearchResult       = "search_result"
	MessageTypeLoading            = "loading"
	MessageTypeGenerativeImage    = "generative_image"
	MessageTypeMessageText        = "message"
	MessageTypeSuggestedResponses = "suggested_responses"
	MessageTypeError              = "error"
)

var (
	ErrMessageRevoke   = errors.New("message revoke detected")
	ErrMessageFiltered = errors.New("message triggered the Bing filter")
)

type Message struct {
	Type  string
	Text  string
	Error error
}
type ChatMessage struct {
	Arguments    []Argument `json:"arguments"`
	InvocationId string     `json:"invocationId"`
	Target       string     `json:"target"`
	Type         int        `json:"type"`
}
type Argument struct {
	OptionsSets           []string          `json:"optionsSets"`
	Source                string            `json:"source"`
	AllowedMessageTypes   []string          `json:"allowedMessageTypes"`
	SliceIds              []string          `json:"sliceIds"`
	Verbosity             string            `json:"verbosity"`
	Scenario              string            `json:"scenario"`
	TraceId               string            `json:"traceId"`
	RequestId             string            `json:"requestId"`
	IsStartOfSession      bool              `json:"isStartOfSession"`
	Message               ArgumentMessage   `json:"message"`
	Tone                  string            `json:"tone"`
	ConversationSignature any               `json:"conversationSignature"`
	Participant           Participant       `json:"participant"`
	SpokenTextMode        string            `json:"spokenTextMode"`
	ConversationId        string            `json:"conversationId"`
	PreviousMessages      []PreviousMessage `json:"previousMessages"`
}
type ArgumentMessage struct {
	Locale        string         `json:"locale"`
	Market        string         `json:"market"`
	Region        string         `json:"region"`
	LocationHints []LocationHint `json:"locationHints"`
	Author        string         `json:"author"`
	InputMethod   string         `json:"inputMethod"`
	Text          string         `json:"text"`
	MessageType   string         `json:"messageType"`
	RequestId     string         `json:"requestId"`
	MessageId     string         `json:"messageId"`
	ImageUrl      any            `json:"imageUrl"`
}
type Participant struct {
	Id string `json:"id"`
}
type PreviousMessage struct {
	Author      string `json:"author"`
	Description string `json:"description"`
	ContextType string `json:"contextType"`
	MessageType string `json:"messageType"`
	MessageId   string `json:"messageId"`
}
type AskStreamOptions struct {
	StopCtx        context.Context
	Conversation   CreateConversationResponse
	Prompt         string
	WebpageContext string
	ImageURL       string
}
