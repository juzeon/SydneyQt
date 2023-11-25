package main

import (
	"context"
	"errors"
	"github.com/pkoukk/tiktoken-go"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"log/slog"
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
}

const (
	ChatFinishResultErrTypeMessageRevoke   = "message_revoke"
	ChatFinishResultErrTypeMessageFiltered = "message_filtered"
	ChatFinishResultErrTypeOthers          = "others"
)

type ChatFinishResult struct {
	Success bool   `json:"success"`
	ErrType string `json:"err_type"`
	ErrMsg  string `json:"err_msg"`
}

const (
	EventConversationCreated    = "chat_conversation_created"
	EventChatAppend             = "chat_append"
	EventChatFinish             = "chat_finish"
	EventChatSuggestedResponses = "chat_suggested_responses"
	EventChatToken              = "chat_token"
)

const (
	EventChatStop = "chat_stop"
)

func (a *App) Dummy1() ChatFinishResult {
	return ChatFinishResult{}
}

func (a *App) askSydney(options AskOptions) {
	chatFinishResult := ChatFinishResult{
		Success: true,
		ErrType: "",
		ErrMsg:  "",
	}
	defer func() {
		slog.Info("invoke EventChatFinish", "result", chatFinishResult)
		runtime.EventsEmit(a.ctx, EventChatFinish, chatFinishResult)
	}()
	currentWorkspace := a.settings.config.GetCurrentWorkspace()
	sydneyIns := sydney.NewSydney(a.debug, util.ReadCookiesFile(), a.settings.config.Proxy,
		currentWorkspace.ConversationStyle, currentWorkspace.Locale, a.settings.config.WssDomain,
		currentWorkspace.NoSearch)
	conversation, err := sydneyIns.CreateConversation()

	if err != nil {
		chatFinishResult = ChatFinishResult{
			Success: false,
			ErrType: ChatFinishResultErrTypeOthers,
			ErrMsg:  err.Error(),
		}
		return
	}
	runtime.EventsEmit(a.ctx, EventConversationCreated)
	stopCtx, cancel := util.CreateCancelContext()
	defer cancel()
	go func() {
		runtime.EventsOn(a.ctx, EventChatStop, func(optionalData ...interface{}) {
			slog.Info("Received EventChatStop")
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
				chatFinishResult = ChatFinishResult{
					Success: false,
					ErrType: ChatFinishResultErrTypeMessageRevoke,
					ErrMsg:  msg.Error.Error(),
				}
			} else if errors.Is(msg.Error, sydney.ErrMessageFiltered) {
				chatFinishResult = ChatFinishResult{
					Success: false,
					ErrType: ChatFinishResultErrTypeMessageFiltered,
					ErrMsg:  msg.Error.Error(),
				}
			} else {
				chatFinishResult = ChatFinishResult{
					Success: false,
					ErrType: ChatFinishResultErrTypeOthers,
					ErrMsg:  msg.Error.Error(),
				}
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
		runtime.EventsEmit(a.ctx, EventChatFinish, ChatFinishResult{
			Success: false,
			ErrType: ChatFinishResultErrTypeOthers,
			ErrMsg:  "not implemented",
		})
	}
}

var tk *tiktoken.Tiktoken
var initTkFunc = sync.OnceFunc(func() {
	slog.Info("Init tiktoken")
	t, err := tiktoken.EncodingForModel("gpt-4")
	if err != nil {
		panic(t)
	}
	tk = t
})

func (a *App) CountToken(text string) int {
	initTkFunc()
	return len(tk.Encode(text, nil, nil))
}
