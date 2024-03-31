package server

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/go-github/v60/github"
	"github.com/ross96D/updater/server/auth"
	"github.com/ross96D/updater/server/github_handler"
	"github.com/ross96D/updater/server/user_handler"
	"github.com/ross96D/updater/share"
)

type Server struct {
	router *chi.Mux
}

func New() *Server {
	s := new(Server)
	s.router = chi.NewMux()
	s.setHandlers()
	return s
}

func (s *Server) Start() error {
	log.Println("starting server on " + ":" + strconv.Itoa(int(share.Config().Port)))
	return http.ListenAndServe(":"+strconv.Itoa(int(share.Config().Port)), s.router)
}

func (s *Server) setHandlers() {
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.DefaultLogger)
	s.router.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddelware)
		r.Post("/update", update)
		r.Get("/list", list)
	})
	s.router.Post("/login", login)
}

func login(w http.ResponseWriter, r *http.Request) {
	name, pass, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}
	valid := false
	for _, user := range share.Config().Users {
		if name == user.Name && pass == user.Password {
			valid = true
			break
		}
	}
	if !valid {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}
	token, err := auth.NewUserToken(name)
	if err != nil {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}
	w.Write(token)
}

func update(w http.ResponseWriter, r *http.Request) {
	origin := r.Context().Value(auth.UserTypeKey)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	defer r.Body.Close()
	switch origin {
	case "github":
		eventType := r.Header.Get(github.EventTypeHeader)
		err = github_handler.HandleGithubWebhook(payload, eventType)
		if err != nil {
			http.Error(w, "internal error", 500)
		}

	case "user":
		err = user_handler.HandlerUserUpdate(payload)
		if err != nil {
			http.Error(w, "internal error", 500)
		}
	default:
		log.Println("unhandled origin")
		http.Error(w, "internal error", 500)
	}
}

func list(w http.ResponseWriter, r *http.Request) {
	origin := r.Context().Value(auth.UserTypeKey)
	if origin != "user" {
		http.Error(w, "", 400)
		return
	}

	err := user_handler.HandleUserAppsList(w)
	if err != nil {
		// log here
		// maybe this is not necesary? this would panic or error or something like that
		http.Error(w, err.Error(), 500)
	}
}
