package server

import (
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-github/v60/github"
	"github.com/ross96D/updater/server/auth"
	"github.com/ross96D/updater/share"
)

type Server struct {
	router *chi.Mux
}

func New() *Server {
	s := new(Server)
	s.router = chi.NewMux()
	return s
}

func (s *Server) Start() error {
	return http.ListenAndServe(":"+strconv.Itoa(int(share.Config().Port)), s.router)
}

func (s *Server) setHandlers() {
	s.router.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddelware)
		r.Get("/update", update)
	})
}

func update(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	eventType := r.Header.Get(github.EventTypeHeader)
	handleGithubWebhook(payload, eventType)
	// call github api to check if we should update

	// if there is an update available
}
