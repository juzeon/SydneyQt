package sydney

import (
	"context"
	"errors"
	"time"
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
	MessageTypeExecutingTask      = "executing_task"
	MessageTypeGeneratedCode      = "generated_code"
	MessageTypeSolvingCaptcha     = "solving_captcha"
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
	GptId                 string            `json:"gptId"`
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
type Options struct {
	Debug                 bool
	Cookies               map[string]string
	Proxy                 string
	ConversationStyle     string
	Locale                string
	WssDomain             string
	CreateConversationURL string
	NoSearch              bool
	GPT4Turbo             bool
	BypassServer          string
}
type AskStreamOptions struct {
	StopCtx        context.Context
	Prompt         string
	WebpageContext string
	ImageURL       string

	MessageID string // A random uuid. Optional.
}
type UploadImagePayload struct {
	ImageInfo        map[string]any   `json:"imageInfo"`
	KnowledgeRequest KnowledgeRequest `json:"knowledgeRequest"`
}
type InvokedSkillsRequestData struct {
	EnableFaceBlur bool `json:"enableFaceBlur"`
}
type ConvoData struct {
	Convoid   string `json:"convoid"`
	Convotone string `json:"convotone"`
}
type KnowledgeRequest struct {
	InvokedSkills            []string                 `json:"invokedSkills"`
	SubscriptionId           string                   `json:"subscriptionId"`
	InvokedSkillsRequestData InvokedSkillsRequestData `json:"invokedSkillsRequestData"`
	ConvoData                ConvoData                `json:"convoData"`
}
type UploadImageResponse struct {
	BlobId          string `json:"blobId"`
	ProcessedBlobId string `json:"processedBlobId"`
}
type GenerativeImage struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}
type GenerateImageResult struct {
	GenerativeImage
	ImageURLs []string      `json:"image_urls"`
	Duration  time.Duration `json:"duration"`
}
type SourceAttribute struct {
	Index int    `json:"index"`
	Link  string `json:"link"`
	Title string `json:"title"`
}
type BypassCaptchaRequest struct {
	IG       string `json:"IG"`
	Cookies  string `json:"cookies"`
	IFrameID string `json:"iframeid"`
	ConvID   string `json:"convId"`
	RID      string `json:"rid"`
}
type BypassCaptchaResponse struct {
	Result struct {
		Cookies    string `json:"cookies"`
		ScreenShot string `json:"screenshot"`
	} `json:"result"`
	Error string `json:"error"`
}
