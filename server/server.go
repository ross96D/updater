package server

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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
	return http.ListenAndServe(":"+strconv.Itoa(share.Config().Port), s.router)
}

func (s *Server) setHandlers() {
	s.router.Group(func(r chi.Router) {
		r.Use(autorizeGithub)
		r.Get("/update", update)
	})
}

func update(w http.ResponseWriter, r *http.Request) {
	// call github api to check if we should update

	// if there is an update available
}
