package server

import (
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ross96D/updater/server/auth"
	"github.com/ross96D/updater/server/user_handler"
	"github.com/ross96D/updater/share"
	"github.com/ross96D/updater/share/configuration"
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
	// TODO change to zerologger and improve the log
	s.router.Use(middleware.DefaultLogger)
	s.router.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddelware)
		r.Post("/update", upload)
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

func list(w http.ResponseWriter, r *http.Request) {
	origin := r.Context().Value(auth.TypeKey)
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
	if r.Context().Value(auth.TypeKey) != "webhook" {
		http.Error(w, "unsupported: "+r.Context().Value(auth.TypeKey).(string), 500)
		return
	}

	app := r.Context().Value(auth.AppValueKey).(configuration.Application)

	data, err := ParseForm(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	err = share.Update(app, data)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	log.Info().Msg("upload update success")
	w.WriteHeader(200)
}

type Data struct {
	form *multipart.Form
}

func (d Data) Get(name string) io.ReadCloser {
	headers, ok := d.form.File[name]
	if !ok || len(headers) == 0 {
		return nil
	}
	header := headers[0]

	r, err := header.Open()
	if err != nil {
		return nil
	}
	return r
}

func ParseForm(r *http.Request) (share.Data, error) {
	err := r.ParseMultipartForm(10 << 20) // store 10 MB in memory
	if err != nil {
		return nil, err
	}

	return Data{form: r.MultipartForm}, nil
}
