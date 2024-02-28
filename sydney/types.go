package sydney

import (
	"context"
	"errors"
	"time"
)

const delimiter = '\x1e'

type LocationHint struct {
	SourceType               int     `json:"SourceType"`
	RegionType               int     `json:"RegionType"`
	Center                   LatLng  `json:"Center"`
	Radius                   int     `json:"Radius"`
	Name                     string  `json:"Name"`
	Accuracy                 int     `json:"Accuracy"`
	FDConfidence             float64 `json:"FDConfidence"`
	CountryName              string  `json:"CountryName"`
	CountryConfidence        int     `json:"CountryConfidence"`
	Admin1Name               string  `json:"Admin1Name"`
	PopulatedPlaceName       string  `json:"PopulatedPlaceName"`
	PopulatedPlaceConfidence int     `json:"PopulatedPlaceConfidence"`
	PostCodeName             string  `json:"PostCodeName"`
	UtcOffset                int     `json:"UtcOffset"`
	Dma                      int     `json:"Dma"`
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
	MessageTypeGenerativeMusic    = "generative_music"
	MessageTypeExecutingTask      = "executing_task"
	MessageTypeOpenAPICall        = "openapi_call"
	MessageTypeGeneratedCode      = "generated_code"
	MessageTypeResolvingCaptcha   = "resolving_captcha"
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
	Plugins               []ArgumentPlugin  `json:"plugins"`
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
type ArgumentPlugin struct {
	Id       string `json:"id"`
	Category int    `json:"category"`
}
type ArgumentMessage struct {
	Locale        string         `json:"locale"`
	Market        string         `json:"market"`
	Region        string         `json:"region"`
	Location      string         `json:"location"`
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
	UseClassic            bool
	GPT4Turbo             bool
	BypassServer          string
}
type AskStreamOptions struct {
	StopCtx        context.Context
	Prompt         string
	WebpageContext string
	ImageURL       string

	messageID            string // A random uuid. Optional.
	disableCaptchaBypass bool
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
type GenerativeMusic struct {
	IFrameID  string `json:"iframeid"`
	RequestID string `json:"requestid"`
	Text      string `json:"text"`
}
type GenerateImageResult struct {
	GenerativeImage
	ImageURLs []string      `json:"image_urls"`
	Duration  time.Duration `json:"duration"`
}
type GenerateMusicResult struct {
	GenerativeMusic
	CoverImgURL   string        `json:"cover_img_url"`
	AudioURL      string        `json:"music_url"`
	VideoURL      string        `json:"video_url"`
	MusicDuration time.Duration `json:"duration"`
	MusicalStyle  string        `json:"musical_style"`
	Title         string        `json:"title"`
	TimeElapsed   time.Duration `json:"time_elapsed"`
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
