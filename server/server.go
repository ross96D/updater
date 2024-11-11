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
	"github.com/ross96D/updater/share/match"
	"github.com/ross96D/updater/share/utils"
	"github.com/ross96D/updater/upgrade"
	"github.com/rs/zerolog/hlog"
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
		r.Get("/config", Config)
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
	err := upgrade.Upgrade(upgrade.Updater)
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
	r.Body.Close()
	os.Exit(1)
}

func Config(w http.ResponseWriter, r *http.Request) {
	data, err := share.ReadConfigFile()
	if err != nil {
		log.Error().Err(err).Send()
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(200)
	_, err = w.Write(data)
	if err != nil {
		log.Error().Err(err).Msg("sending config file")
	}
}

func Reload(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Send()
		http.Error(w, err.Error(), 500)
		return
	}
	if len(data) == 0 {
		http.Error(w, "invalid: emtpy config", 500)
		return
	}

	err = share.ReloadString(string(data))
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	err = share.ReplaceConfigFile(data)
	if err != nil {
		log.Error().Err(err).Msg("replacing config file")
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
	logger := logger.ResponseWithLogger.FromContext(r.Context())

	dryRun := r.Header.Get("dry-run") == "true"

	switch r.Context().Value(auth.TypeKey) {
	case "webhook":
		app := r.Context().Value(auth.AppValueKey).(configuration.Application)

		data, err := ParseForm(r)
		// we need to parse the body first before sending a message
		logger.Info().Bool("dry-run", dryRun).Send()
		if err != nil {
			logger.Error().Err(err).Msg("ParseForm")
			return
		}

		err = match.Update(r.Context(), app, match.WithData(data), match.WithDryRun(dryRun))

		if err != nil {
			switch err.(type) {
			case match.ErrErrors, match.ErrError:
				logger.Error().Err(err).
					Str("reqID", utils.Ignore2(hlog.IDFromCtx(r.Context())).String()).
					Send()
			default:
				logger.Warn().Err(err).
					Str("reqID", utils.Ignore2(hlog.IDFromCtx(r.Context())).String()).
					Send()
			}
			return
		}
	case "user":
		payload, err := io.ReadAll(r.Body)
		// we need to parse the body first before sending a message
		logger.Info().Bool("dry-run", dryRun).Send()
		if err != nil {
			logger.Error().Err(err).Msg("reading data")
			return
		}
		defer r.Body.Close()

		err = user_handler.HandlerUserUpdate(r.Context(), payload, dryRun)

		if err != nil {
			switch err.(type) {
			// TODO needs to refactor this because all real errors would need to be of this type
			case match.ErrErrors, match.ErrError:
				logger.Error().Err(err).
					Str("reqID", utils.Ignore2(hlog.IDFromCtx(r.Context())).String()).
					Send()
			default:
				logger.Warn().Err(err).
					Str("reqID", utils.Ignore2(hlog.IDFromCtx(r.Context())).String()).
					Send()
			}
			return
		}
	default:
		http.Error(w, "unsupported: "+r.Context().Value(auth.TypeKey).(string), 500)
		return
	}

	logger.Info().Msg("upload update success")
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

func ParseForm(r *http.Request) (match.Data, error) {
	err := r.ParseMultipartForm(10 << 20) // store 10 MB in memory
	if err != nil {
		df := make([]byte, 100)
		n, err2 := r.Body.Read(df)
		return nil, fmt.Errorf("%d %s %w", n, err2.Error(), err)
	}

	return Data{form: r.MultipartForm}, nil
}
