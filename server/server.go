package server

import (
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ross96D/updater/server/auth"
	"github.com/ross96D/updater/server/user_handler"
	"github.com/ross96D/updater/share"
	"github.com/rs/zerolog/log"
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
	log.Info().Msg("starting server on " + ":" + strconv.Itoa(int(share.Config().Port)))
	return http.ListenAndServe(":"+strconv.Itoa(int(share.Config().Port)), s.router)
}

func (s *Server) setHandlers() {
	s.router.Use(middleware.Recoverer)
	// TODO change to zerologger
	s.router.Use(middleware.DefaultLogger)
	s.router.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddelware)
		r.Post("/update", update)
		r.Get("/list", list)
		r.Post("/reload", reload)
	})
	s.router.Post("/login", login)
}

func reload(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Send()
		http.Error(w, err.Error(), 500)
		return
	}

	// TODO is this right? i am not shure because this is only usefull on testing.. maybe use a header to confirm that we want to use the default config file
	if len(data) == 0 {
		err = share.Reload("config.cue")
	} else {
		err = share.ReloadString(string(data))
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
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
	_, _ = w.Write(token)
}

func update(w http.ResponseWriter, r *http.Request) {
	origin := r.Context().Value(auth.UserTypeKey)

	payload, err := io.ReadAll(r.Body)
	_ = payload
	if err != nil {
		log.Error().Err(err).Send()
		http.Error(w, "internal error", 500)
		return
	}
	defer r.Body.Close()
	switch origin {
	case "user":
		log.Warn().Msg("update from user not currently supported")
		http.Error(w, "update from user not currently supported", 400)
		return
	default:
		log.Warn().Msg("unhandled origin")
		http.Error(w, "internal error", 500)
		return
	}
	log.Info().Msg("update success")
	w.WriteHeader(200)
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

func upload(w http.ResponseWriter, r *http.Request) {
	origin := r.Context().Value(auth.UserTypeKey).(string)
	if origin != "github" {
		http.Error(w, "invalid origin "+origin, 403)
	}

	log.Info().Msg("upload update success")
	w.WriteHeader(200)
}
