package logger

import (
	"context"
	"net/http"
	"time"

	"github.com/justinas/alice"
	"github.com/ross96D/updater/share/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

var LoggerMiddleware = func(next http.Handler) http.Handler {
	c := alice.New(
		hlog.NewHandler(log.Logger),
		hlog.ProtoHandler("proto"),
		hlog.RequestHandler("url"),
		hlog.RemoteAddrHandler("ip"),
		hlog.UserAgentHandler("user_agent"),
		hlog.RequestIDHandler("req_id", "Request-Id"),
		hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Send()
		}),
	)

	return c.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	}))
}

type loggerCtxKey struct{}
type LoggerCtx func(next http.Handler) http.Handler

func LoggerCtx_WithContex(ctx context.Context, logger *zerolog.Logger, handler *Handler) context.Context {
	return context.WithValue(ctx, loggerCtxKey{}, &struct {
		logger  *zerolog.Logger
		handler *Handler
	}{logger: logger, handler: handler})
}

var nop = zerolog.Nop()

func LoggerCtx_FromContext(ctx context.Context) (*zerolog.Logger, *Handler) {
	if r, ok := ctx.Value(loggerCtxKey{}).(*struct {
		logger  *zerolog.Logger
		handler *Handler
	}); ok {
		return r.logger, r.handler
	} else {
		panic("responseWithLogger FromContext called without a value")
	}
}

var ResponseWithLogger LoggerCtx = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler, logger, err := New(log.Logger, r, w)
		utils.Assert(err == nil, "could not create logger %s", err)

		r = r.WithContext(LoggerCtx_WithContex(r.Context(), logger, handler))
		next.ServeHTTP(w, r)
	})
}
