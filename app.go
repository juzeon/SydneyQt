package main

import (
	"context"
	"errors"
	"github.com/pkoukk/tiktoken-go"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"sydneyqt/sydney"
	"sydneyqt/util"
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
	EventConversationCreated    = "chat_conversation_created"
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
	sydneyIns := sydney.NewSydney(a.debug, util.ReadCookiesFile(), a.settings.config.Proxy,
		currentWorkspace.ConversationStyle, currentWorkspace.Locale, a.settings.config.WssDomain,
		a.settings.config.NoSearch)
	conversation, err := sydneyIns.CreateConversation()
	if err != nil {
		runtime.EventsEmit(a.ctx, EventChatAlert, err.Error())
		return
	}
	runtime.EventsEmit(a.ctx, EventConversationCreated)
	stopCtx, cancel := util.CreateCancelContext()
	defer cancel()
	go func() {
		runtime.EventsOn(a.ctx, EventChatStop, func(optionalData ...interface{}) {
			cancel()
		})
	}()
	ch := sydneyIns.AskStream(sydney.AskStreamOptions{
		StopCtx:        stopCtx,
		Conversation:   conversation,
		Prompt:         options.Prompt,
		WebpageContext: options.ChatContext,
		ImageURL:       options.ImageURL,
	})
	defer runtime.EventsEmit(a.ctx, EventChatFinish)
	chatAppend := func(text string) {
		runtime.EventsEmit(a.ctx, EventChatAppend, text)
	}
	fullMessageText := ""
	lastMessageType := ""
	for msg := range ch {
		textToAppend := ""
		switch msg.Type {
		case sydney.MessageTypeSuggestedResponses:
			runtime.EventsEmit(a.ctx, EventChatSuggestedResponses, msg.Text)
		case sydney.MessageTypeError:
			if errors.Is(msg.Error, sydney.ErrMessageRevoke) {
				runtime.EventsEmit(a.ctx, EventChatMessageRevoke, options.ReplyDeep)
				if a.settings.config.RevokeReplyText == "" || options.ReplyDeep >= a.settings.config.RevokeReplyCount {
					runtime.EventsEmit(a.ctx, EventChatAlert, msg.Text)
				}
			} else {
				runtime.EventsEmit(a.ctx, EventChatAlert, msg.Text)
			}
			return
		case sydney.MessageTypeMessageText:
			fullMessageText += msg.Text
			runtime.EventsEmit(a.ctx, EventChatToken, a.CountToken(fullMessageText))
			textToAppend = msg.Text
		default:
			textToAppend = msg.Text + "\n\n"
		}
		if textToAppend != "" {
			if lastMessageType != msg.Type {
				textToAppend = "[assistant](#" + msg.Type + ")\n" + textToAppend
			}
			chatAppend(textToAppend)
		}
		lastMessageType = msg.Type
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
