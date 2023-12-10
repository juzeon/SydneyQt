package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/life4/genesis/slices"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"io"
	"log/slog"
	"sydneyqt/sydney"
	"sydneyqt/util"
)

const (
	_ = iota
	AskTypeSydney
	AskTypeOpenAI
)

type AskType int

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
	EventChatGenerateImage      = "chat_generate_image"
)

const (
	EventChatStop = "chat_stop"
)

func (a *App) Dummy1() ChatFinishResult {
	return ChatFinishResult{}
}
func (a *App) createSydney() (*sydney.Sydney, error) {
	currentWorkspace, err := a.settings.config.GetCurrentWorkspace()
	if err != nil {
		return nil, err
	}
	cookies, err := util.ReadCookiesFile()
	if err != nil {
		return nil, err
	}
	return sydney.NewSydney(a.debug, cookies, a.settings.config.Proxy,
		currentWorkspace.ConversationStyle, currentWorkspace.Locale, a.settings.config.WssDomain,
		a.settings.config.CreateConversationURL,
		currentWorkspace.NoSearch), nil
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
	sydneyIns, err := a.createSydney()
	if err != nil {
		chatFinishResult = ChatFinishResult{
			Success: false,
			ErrType: ChatFinishResultErrTypeOthers,
			ErrMsg:  err.Error(),
		}
		return
	}
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
		case sydney.MessageTypeGenerativeImage:
			var image sydney.GenerativeImage
			lo.Must0(json.Unmarshal([]byte(msg.Text), &image))
			runtime.EventsEmit(a.ctx, EventChatGenerateImage, image)
			textToAppend = image.Text + "\n\n"
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
func (a *App) askOpenAI(options AskOptions) {
	chatFinishResult := ChatFinishResult{
		Success: true,
		ErrType: "",
		ErrMsg:  "",
	}
	handleErr := func(err error) {
		chatFinishResult = ChatFinishResult{
			Success: false,
			ErrType: ChatFinishResultErrTypeOthers,
			ErrMsg:  err.Error(),
		}
	}
	defer func() {
		slog.Info("invoke EventChatFinish", "result", chatFinishResult)
		runtime.EventsEmit(a.ctx, EventChatFinish, chatFinishResult)
	}()
	backend, err := slices.Find(a.settings.config.OpenAIBackends, func(el OpenAIBackend) bool {
		return el.Name == options.OpenAIBackend
	})
	if err != nil {
		handleErr(err)
		return
	}
	slog.Info("Ask OpenAI with backend: ", "data", backend)
	hClient, err := util.MakeHTTPClient(a.settings.config.Proxy, 0)
	if err != nil {
		handleErr(err)
		return
	}
	config := openai.DefaultConfig(backend.OpenaiKey)
	config.BaseURL = backend.OpenaiEndpoint
	config.HTTPClient = hClient
	client := openai.NewClientWithConfig(config)
	messages := util.GetOpenAIChatMessages(options.ChatContext)
	slog.Info("Get chat messages", "messages", messages)
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: options.Prompt,
	})
	stream, err := client.CreateChatCompletionStream(context.Background(), openai.ChatCompletionRequest{
		Model:            backend.OpenaiShortModel,
		Messages:         messages,
		Temperature:      backend.OpenaiTemperature,
		FrequencyPenalty: backend.FrequencyPenalty,
		PresencePenalty:  backend.PresencePenalty,
		MaxTokens:        backend.MaxTokens,
	})
	if err != nil {
		handleErr(err)
		return
	}
	runtime.EventsEmit(a.ctx, EventConversationCreated)
	defer stream.Close()
	fullMessage := ""
	replied := false
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			slog.Info("openai chat completed")
			return
		}
		if err != nil {
			handleErr(err)
			return
		}
		textToAppend := response.Choices[0].Delta.Content
		fullMessage += textToAppend
		runtime.EventsEmit(a.ctx, EventChatToken, a.CountToken(fullMessage))
		if !replied {
			textToAppend = "[assistant](#message)\n" + textToAppend
			replied = true
		}
		runtime.EventsEmit(a.ctx, EventChatAppend, textToAppend)
	}
}
func (a *App) AskAI(options AskOptions) {
	if options.Type == AskTypeSydney {
		a.askSydney(options)
	} else if options.Type == AskTypeOpenAI {
		a.askOpenAI(options)
	}
}
