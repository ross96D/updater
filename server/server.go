package server

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ross96D/updater/logger"
	"github.com/ross96D/updater/server/auth"
	"github.com/ross96D/updater/server/user_handler"
	"github.com/ross96D/updater/share"
	"github.com/ross96D/updater/share/configuration"
	"github.com/ross96D/updater/upgrade"
	"github.com/rs/zerolog/log"
)

type Server struct {
	router   *chi.Mux
	keyPath  string
	certPath string
}

func New(keyPath, certPath string) *Server {
	s := new(Server)
	s.certPath = certPath
	s.keyPath = keyPath
	s.router = chi.NewMux()
	s.setHandlers()
	return s
}

func (s *Server) TestServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) Start() error {
	log.Info().Msg("starting server on " + ":" + strconv.Itoa(int(share.Config().Port)))
	portStr := ":" + strconv.Itoa(int(share.Config().Port))
	if s.certPath != "" && s.keyPath != "" {
		return http.ListenAndServeTLS(portStr, s.certPath, s.keyPath, s.router)
	} else {
		return http.ListenAndServe(portStr, s.router)
	}
}

func (s *Server) setHandlers() {
	s.router.Use(middleware.Recoverer)
	s.router.Use(logger.LoggerMiddleware)
	s.router.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddelware)
		r.Get("/list", List)
		r.Group(func(r chi.Router) {
			r.Use(logger.ResponseWithLogger)
			r.Post("/update", Update)
		})
		r.Post("/reload", Reload)
		r.Post("/upgrade", Upgrade)
	})
	s.router.Post("/login", Login)
}

func Upgrade(w http.ResponseWriter, r *http.Request) {
	if r.Context().Value(auth.TypeKey) != "user" {
		http.Error(w, "", 403)
		return
	}
	err := upgrade.Upgrade()
	if err == upgrade.ErrUpToDate {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_, err = w.Write([]byte("success. quiting process"))
	if err != nil {
		log.Warn().Err(fmt.Errorf("upgradeUpdater %w", err)).Send()
	}
	err = http.NewResponseController(w).Flush()
	if err != nil {
		log.Warn().Err(fmt.Errorf("upgradeUpdater %w", err)).Msg("flushing response")
	}
	os.Exit(1)
}

func Reload(w http.ResponseWriter, r *http.Request) {
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

func Login(w http.ResponseWriter, r *http.Request) {
	name, pass, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "No basic auth", http.StatusUnauthorized)
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
		http.Error(w, "invalid authentication", http.StatusUnauthorized)
		return
	}
	token, err := auth.NewUserToken(name)
	if err != nil {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}
	_, _ = w.Write(token)
}

func List(w http.ResponseWriter, r *http.Request) {
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

func Update(w http.ResponseWriter, r *http.Request) {
	switch r.Context().Value(auth.TypeKey) {
	case "webhook":
		app := r.Context().Value(auth.AppValueKey).(configuration.Application)

		data, err := ParseForm(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = share.Update(r.Context(), app, share.WithData(data))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	case "user":
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Send()
			http.Error(w, "internal error", 500)
			return
		}
		defer r.Body.Close()

		err = user_handler.HandlerUserUpdate(r.Context(), payload)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	default:
		http.Error(w, "unsupported: "+r.Context().Value(auth.TypeKey).(string), 500)
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
