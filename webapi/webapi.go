package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sydneyqt/sydney"
	"sydneyqt/util"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// read envs
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	proxy := os.Getenv("HTTPS_PROXY")
	if proxy == "" {
		proxy = os.Getenv("HTTP_PROXY")
	}

	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "*"
	}

	noLog := os.Getenv("NO_LOG") != ""

	defaultCookies := ParseCookies(os.Getenv("DEFAULT_COOKIES"))

	authToken := os.Getenv("AUTH_TOKEN")

	// create router
	r := chi.NewRouter()

	// set middleware
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	if !noLog {
		r.Use(middleware.Logger)
	}
	r.Use(middleware.SetHeader("Access-Control-Allow-Origin", allowedOrigins))
	// auth middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authToken == "" {
				next.ServeHTTP(w, r)
				return
			}
			if r.Header.Get("Authorization") != "Bearer "+authToken {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// add handlers
	r.Post("/conversation/new", func(w http.ResponseWriter, r *http.Request) {
		// parse request
		var request CreateConversationRequest

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cookies := util.Ternary(request.Cookies == "", defaultCookies, ParseCookies(request.Cookies))

		// create conversation
		conversation, err := sydney.
			NewSydney(false, cookies, proxy, "", "", "", "", false).
			CreateConversation()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// set headers
		w.Header().Set("Content-Type", "application/json")

		// write response
		json.NewEncoder(w).Encode(conversation)
	})

	r.Post("/image/upload", func(w http.ResponseWriter, r *http.Request) {
		// parse request
		r.ParseMultipartForm(16 << 20)

		cookiesStr := r.FormValue("cookies")
		cookies := util.Ternary(cookiesStr == "", defaultCookies, ParseCookies(cookiesStr))

		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		bytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// upload image
		imgUrl, err := sydney.
			NewSydney(false, cookies, proxy, "Creative", "", "", "", false).
			UploadImage(bytes)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// set headers
		w.Header().Set("Content-Type", "text/plain")

		// write response
		fmt.Fprint(w, imgUrl)
	})

	r.Post("/chat/stream", func(w http.ResponseWriter, r *http.Request) {
		// parse request
		request := ChatStreamRequest{
			ConversationStyle: "Creative",
		}

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cookies := util.Ternary(request.Cookies == "", defaultCookies, ParseCookies(request.Cookies))

		// avoid panic caused by invalid locale
		if len(request.Locale) != 5 {
			request.Locale = "en-US"
		}

		// stream chat
		messageCh := sydney.
			NewSydney(false, cookies, proxy, request.ConversationStyle, request.Locale, "", "", request.NoSearch).
			AskStream(sydney.AskStreamOptions{
				StopCtx:        r.Context(),
				Conversation:   request.Conversation,
				Prompt:         request.Prompt,
				WebpageContext: request.WebpageContext,
				ImageURL:       request.ImageURL,
			})

		// set headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// write response
		for message := range messageCh {
			encoded, _ := json.Marshal(message.Text)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", message.Type, encoded)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	})

	r.Post("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		// parse request
		var request OpenAIChatCompletionRequest

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		parsedMessages, err := ParseOpenAIMessages(request.Messages)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cookiesStr := r.Header.Get("Cookie")
		cookies := util.Ternary(cookiesStr == "", defaultCookies, ParseCookies(cookiesStr))

		conversationStyle := util.Ternary(
			strings.HasPrefix(request.Model, "gpt-3.5-turbo"), "Balanced", "Creative")

		sydneyAPI := sydney.NewSydney(false, cookies, proxy, conversationStyle, "en-US", "", "", false)

		// create new conversation if not provided
		if request.Conversation.ConversationId == "" {
			request.Conversation, err = sydneyAPI.CreateConversation()
			if err != nil {
				http.Error(w, "error creating conversation: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		messageCh := sydneyAPI.AskStream(sydney.AskStreamOptions{
			StopCtx:        r.Context(),
			Conversation:   request.Conversation,
			Prompt:         parsedMessages.Prompt,
			WebpageContext: parsedMessages.WebpageContext,
			ImageURL:       parsedMessages.ImageURL,
		})

		// set headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// write response
		for message := range messageCh {
			encoded, _ := json.Marshal(message.Text)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", message.Type, encoded)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	})

	// serve the router
	log.Println("Listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
