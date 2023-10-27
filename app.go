package main

import (
	"context"
	"fmt"
	"github.com/pkoukk/tiktoken-go"
	"github.com/tidwall/gjson"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"log"
	"strings"
	"sync"
)

const (
	_ = iota
	AskTypeSydney
	AskTypeOpenAI
)

type AskType int

// App struct
type App struct {
	debug    bool
	settings *Settings
	ctx      context.Context
}

// NewApp creates a new App application struct
func NewApp(settings *Settings) *App {
	return &App{debug: false, settings: settings}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

type AskOptions struct {
	Type          AskType `json:"type"`
	OpenAIBackend string  `json:"openai_backend"`
	ChatContext   string  `json:"chat_context"`
	Prompt        string  `json:"prompt"`
	ImageURL      string  `json:"image_url"`
	ReplyDeep     int     `json:"reply_deep"`
}

const (
	EventChatAlert              = "chat_alert"
	EventChatAppend             = "chat_append"
	EventChatFinish             = "chat_finish"
	EventChatSuggestedResponses = "chat_suggested_responses"
	EventChatToken              = "chat_token"
	EventChatMessageRevoke      = "chat_message_revoke"
)

const (
	EventChatStop = "chat_stop"
)

func (a *App) askSydney(options AskOptions) {
	currentWorkspace := a.settings.config.GetCurrentWorkspace()
	sydney := NewSydney(a.debug, ReadCookiesFile(), a.settings.config.Proxy,
		currentWorkspace.ConversationStyle, currentWorkspace.Locale, a.settings.config.WssDomain,
		a.settings.config.NoSearch)
	conversation, err := sydney.CreateConversation()
	if err != nil {
		runtime.EventsEmit(a.ctx, EventChatAlert, err.Error())
		return
	}
	stopCtx, cancel := CreateCancelContext()
	defer cancel()
	go func() {
		runtime.EventsOn(a.ctx, EventChatStop, func(optionalData ...interface{}) {
			cancel()
		})
	}()
	ch := sydney.AskStream(stopCtx, conversation, options.Prompt, options.ChatContext, options.ImageURL)
	defer runtime.EventsEmit(a.ctx, EventChatFinish)
	sendSuggestedResponses := func(message gjson.Result) {
		if message.Get("suggestedResponses").Exists() {
			runtime.EventsEmit(a.ctx, EventChatSuggestedResponses,
				Map(message.Get("suggestedResponses").Array(), func(v gjson.Result) string {
					return v.Get("text").String()
				}),
			)
		}
	}
	chatAppend := func(text string) {
		runtime.EventsEmit(a.ctx, EventChatAppend, text)
	}
	wrote := 0
	replied := false
	for msg := range ch {
		if msg.Error != nil {
			log.Println("error: " + msg.Error.Error())
			runtime.EventsEmit(a.ctx, EventChatAlert, msg.Error.Error())
			return
		}
		data := gjson.Parse(msg.Data)
		if data.Get("type").Int() == 1 && data.Get("arguments.0.messages").Exists() {
			message := data.Get("arguments.0.messages.0")
			msgType := message.Get("messageType")
			messageText := message.Get("text").String()
			messageHiddenText := message.Get("hiddenText").String()
			switch msgType.String() {
			case "InternalSearchQuery":
				chatAppend("[assistant](#search_query)\n" + messageHiddenText + "\n\n")
			case "InternalSearchResult":
				var links []string
				if strings.Contains(messageHiddenText,
					"Web search returned no relevant result") {
					chatAppend("[assistant](#search_query)\n" + messageHiddenText + "\n\n")
					continue
				}
				if !gjson.Valid(messageText) {
					log.Println("Error when parsing InternalSearchResult: " + messageText)
					continue
				}
				arr := gjson.Parse(messageText).Array()
				for _, group := range arr {
					srIndex := 1
					for _, subGroup := range group.Array() {
						links = append(links, fmt.Sprintf("[^%d^][%s](%s)",
							srIndex, subGroup.Get("title").String(), subGroup.Get("url").String()))
						srIndex++
					}
				}
				chatAppend("[assistant](#search_results)\n" + strings.Join(links, "\n\n") + "\n\n")
			case "InternalLoaderMessage":
				if message.Get("hiddenText").Exists() {
					chatAppend("[assistant](#loading)\n" + messageHiddenText + "\n\n")
					continue
				}
				if message.Get("text").Exists() {
					chatAppend("[assistant](#loading)\n" + messageText + "\n\n")
					continue
				}
				chatAppend("[assistant](#loading)\n" + message.Raw + "\n\n")
			case "GenerateContentQuery":
				if message.Get("contentType").String() != "IMAGE" {
					continue
				}
				chatAppend("[assistant](#generative_image)\nKeyword: " +
					messageText + "\n\n")
			case "":
				if data.Get("arguments.0.cursor").Exists() {
					chatAppend("[assistant](#message)\n")
					wrote = 0
				}
				if message.Get("contentOrigin").String() == "Apology" {
					runtime.EventsEmit(a.ctx, EventChatMessageRevoke, options.ReplyDeep)
					if replied &&
						(a.settings.config.RevokeReplyText == "" || options.ReplyDeep >= a.settings.config.RevokeReplyCount) {
						runtime.EventsEmit(a.ctx, EventChatAlert, "Message revoke detected")
					} else {
						runtime.EventsEmit(a.ctx, EventChatAlert,
							"Looks like the user's message has triggered the Bing filter")
					}
					return
				} else {
					replied = true
					chatAppend(messageText[wrote:])
					wrote = len(messageText)
					runtime.EventsEmit(a.ctx, EventChatToken, a.CountToken(messageText))
					sendSuggestedResponses(message)
				}
			default:
				log.Println("Unsupported message type: " + msgType.String())
				log.Println("Triggered by " + options.Prompt + ", response: " + message.Raw)
			}
		} else if data.Get("type").Int() == 2 && data.Get("item.messages").Exists() {
			message := data.Get("item.messages|@reverse|0")
			sendSuggestedResponses(message)
		}
	}
}
func (a *App) AskAI(options AskOptions) {
	if options.Type == AskTypeSydney {
		a.askSydney(options)
	} else if options.Type == AskTypeOpenAI {
		runtime.EventsEmit(a.ctx, EventChatAlert, "not implemented")
	}
}

var tk *tiktoken.Tiktoken

func (a *App) CountToken(text string) int {
	sync.OnceFunc(func() {
		t, err := tiktoken.EncodingForModel("gpt-4")
		if err != nil {
			panic(err)
		}
		tk = t
	})()
	return len(tk.Encode(text, nil, nil))
}
