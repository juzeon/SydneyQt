package main

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/pkoukk/tiktoken-go"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"log/slog"
	"os"
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
func (a *App) createSydney() *sydney.Sydney {
	currentWorkspace := a.settings.config.GetCurrentWorkspace()
	return sydney.NewSydney(a.debug, util.ReadCookiesFile(), a.settings.config.Proxy,
		currentWorkspace.ConversationStyle, currentWorkspace.Locale, a.settings.config.WssDomain,
		currentWorkspace.NoSearch)
}

func (a *App) askSydney(options AskOptions) {
	slog.Info("askSydney called", "options", options)
	chatFinishResult := ChatFinishResult{
		Success: true,
		ErrType: "",
		ErrMsg:  "",
	}
	defer func() {
		slog.Info("invoke EventChatFinish", "result", chatFinishResult)
		runtime.EventsEmit(a.ctx, EventChatFinish, chatFinishResult)
	}()
	sydneyIns := a.createSydney()
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
	runtime.EventsOn(a.ctx, EventChatStop, func(optionalData ...interface{}) {
		slog.Info("Received EventChatStop")
		cancel()
	})
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

type UploadSydneyImageResult struct {
	Base64URL string `json:"base64_url"`
	BingURL   string `json:"bing_url"`
	Canceled  bool   `json:"canceled"`
}

func (a *App) UploadSydneyImage() (UploadSydneyImageResult, error) {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open an image to upload",
		Filters: []runtime.FileFilter{{
			DisplayName: "Image Files (*.jpg; *.jpeg; *.png; *.gif)",
			Pattern:     "*.jpg;*.jpeg;*.png;*.gif",
		}},
	})
	if err != nil {
		return UploadSydneyImageResult{}, err
	}
	if file == "" {
		return UploadSydneyImageResult{Canceled: true}, nil
	}
	sydneyIns := a.createSydney()
	v, err := os.ReadFile(file)
	if err != nil {
		return UploadSydneyImageResult{}, err
	}
	jpgData, err := util.ConvertImageToJpg(v)
	if err != nil {
		return UploadSydneyImageResult{}, err
	}
	url, err := sydneyIns.UploadImage(jpgData)
	if err != nil {
		return UploadSydneyImageResult{}, err
	}
	return UploadSydneyImageResult{
		Base64URL: "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(jpgData),
		BingURL:   url,
	}, err
}
