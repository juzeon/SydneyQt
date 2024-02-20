package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/life4/genesis/slices"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"io"
	"log/slog"
	"strings"
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
	EventChatResolvingCaptcha   = "chat_resolving_captcha"
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
	return sydney.NewSydney(sydney.Options{
		Debug:                 a.settings.config.Debug,
		Cookies:               cookies,
		Proxy:                 a.settings.config.Proxy,
		ConversationStyle:     currentWorkspace.ConversationStyle,
		Locale:                currentWorkspace.Locale,
		WssDomain:             a.settings.config.WssDomain,
		CreateConversationURL: a.settings.config.CreateConversationURL,
		NoSearch:              currentWorkspace.NoSearch,
		UseClassic:            currentWorkspace.UseClassic,
		GPT4Turbo:             currentWorkspace.GPT4Turbo,
		BypassServer:          a.settings.config.BypassServer,
	}), nil
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

	stopCtx, cancel := util.CreateCancelContext()
	defer cancel()
	runtime.EventsOn(a.ctx, EventChatStop, func(optionalData ...interface{}) {
		slog.Info("Received EventChatStop")
		cancel()
		runtime.EventsOff(a.ctx, EventChatStop)
	})

	ch, err := sydneyIns.AskStream(sydney.AskStreamOptions{
		StopCtx:        stopCtx,
		Prompt:         options.Prompt,
		WebpageContext: options.ChatContext,
		ImageURL:       options.ImageURL,
	})
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			chatFinishResult = ChatFinishResult{
				Success: false,
				ErrType: ChatFinishResultErrTypeOthers,
				ErrMsg:  err.Error(),
			}
		}
		return
	}
	runtime.EventsEmit(a.ctx, EventConversationCreated)

	chatAppend := func(text string) {
		runtime.EventsEmit(a.ctx, EventChatAppend, text)
	}
	fullMessageText := ""
	lastMessageType := ""
	receivedBingSearchDisabledLoader := false
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
		case sydney.MessageTypeLoading:
			if a.settings.config.DisableNoSearchLoader {
				if msg.Text == "BingSearchDisabled" {
					receivedBingSearchDisabledLoader = true
					break
				}
				if msg.Text == "Generating answers for you..." && receivedBingSearchDisabledLoader {
					receivedBingSearchDisabledLoader = false
					break
				}
			}
			textToAppend = msg.Text + "\n\n"
		case sydney.MessageTypeResolvingCaptcha:
			runtime.EventsEmit(a.ctx, EventChatResolvingCaptcha, msg.Text)
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
	stopCtx, cancel := context.WithCancel(context.Background())
	runtime.EventsOn(a.ctx, EventChatStop, func(optionalData ...interface{}) {
		slog.Info("Received EventChatStop")
		cancel()
		runtime.EventsOff(a.ctx, EventChatStop)
	})
	backend, err := slices.Find(a.settings.config.OpenAIBackends, func(el OpenAIBackend) bool {
		return el.Name == options.OpenAIBackend
	})
	if err != nil {
		handleErr(err)
		return
	}
	slog.Info("Ask OpenAI with backend: ", "data", backend)
	client, err := util.CreateOpenAIClient(a.settings.config.Proxy, backend.OpenaiKey, backend.OpenaiEndpoint)
	if err != nil {
		handleErr(err)
		return
	}
	messages := util.GetOpenAIChatMessages(options.ChatContext)
	slog.Info("Get chat messages", "messages", messages)
	if options.ImageURL == "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    "user",
			Content: options.Prompt,
		})
	} else {
		messages = append(messages, openai.ChatCompletionMessage{
			Role: "user",
			MultiContent: []openai.ChatMessagePart{{
				Type: openai.ChatMessagePartTypeText,
				Text: options.Prompt,
			}, {
				Type: openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{
					URL: options.ImageURL,
				},
			}},
		})
	}
	stream, err := client.CreateChatCompletionStream(stopCtx, openai.ChatCompletionRequest{
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
		if errors.Is(err, context.Canceled) {
			return
		}
		if err != nil {
			handleErr(err)
			return
		}
		slog.Debug("Received OpenAI delta", "v", response)
		if len(response.Choices) == 0 {
			continue
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

type ConciseAnswerReq struct {
	Prompt  string `json:"prompt"`
	Context string `json:"context"`
	Backend string `json:"backend"`
}

func (a *App) GetConciseAnswer(req ConciseAnswerReq) (string, error) {
	if req.Backend == "Sydney" {
		cookies, err := util.ReadCookiesFile()
		if err != nil {
			return "", err
		}
		syd := sydney.NewSydney(sydney.Options{
			Debug:                 false,
			Cookies:               cookies,
			Proxy:                 a.settings.config.Proxy,
			WssDomain:             a.settings.config.WssDomain,
			CreateConversationURL: a.settings.config.CreateConversationURL,
			NoSearch:              true,
			UseClassic:            false,
		})
		ch, err := syd.AskStream(sydney.AskStreamOptions{
			StopCtx:        context.Background(),
			Prompt:         req.Prompt,
			WebpageContext: req.Context,
		})
		if err != nil {
			return "", err
		}
		var result bytes.Buffer
		for msg := range ch {
			if msg.Type == sydney.MessageTypeError {
				return "", msg.Error
			}
			if msg.Type != sydney.MessageTypeMessageText {
				continue
			}
			result.WriteString(msg.Text)
		}
		return strings.TrimSpace(result.String()), nil
	}
	// openai backends
	backend, ok := lo.Find(a.settings.config.OpenAIBackends, func(item OpenAIBackend) bool {
		return item.Name == req.Backend
	})
	if !ok {
		return "", errors.New("openai backend not found: " + req.Backend)
	}
	client, err := util.CreateOpenAIClient(a.settings.config.Proxy, backend.OpenaiKey, backend.OpenaiEndpoint)
	if err != nil {
		return "", err
	}
	messages := lo.Ternary(req.Context == "", []openai.ChatCompletionMessage{
		{Role: "user", Content: req.Prompt},
	}, []openai.ChatCompletionMessage{
		{Role: "system", Content: req.Context},
		{Role: "user", Content: req.Prompt},
	})
	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:       backend.OpenaiShortModel,
		Messages:    messages,
		Temperature: 1,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("openai len(choices) == 0")
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}
