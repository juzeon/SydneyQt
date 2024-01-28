package main

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"log/slog"
	"net/http"
	"os"
	"sydneyqt/util"
)

type IPCServer struct {
	mux      *chi.Mux
	settings *Settings
}

func NewIPCServer(settings *Settings) *IPCServer {
	mux := chi.NewRouter()
	server := &IPCServer{
		mux:      mux,
		settings: settings,
	}
	server.registerRouters(mux)
	return server
}
func (o *IPCServer) registerRouters(mux *chi.Mux) {
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(cors.Handler(cors.Options{
		AllowOriginFunc: func(r *http.Request, origin string) bool { return true },
		AllowedMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:  []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	}))
	mux.Post("/cookies", func(writer http.ResponseWriter, request *http.Request) {
		var cookies []util.FileCookie
		err := json.NewDecoder(request.Body).Decode(&cookies)
		if err != nil {
			writer.WriteHeader(400)
			slog.Error("Could not decode request", "err", err)
			return
		}
		v, _ := json.MarshalIndent(&cookies, "", "  ")
		err = os.WriteFile("cookies.json", v, 0644)
		if err != nil {
			writer.WriteHeader(500)
			slog.Error("Could write cookies.json", "err", err)
			return
		}
		writer.WriteHeader(200)
	})
}
func (o *IPCServer) Serve() {
	err := http.ListenAndServe(":61989", o.mux)
	if err != nil {
		util.GracefulPanic(err)
	}
}
