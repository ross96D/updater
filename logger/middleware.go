package logger

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/justinas/alice"
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

type responseWithLoggerKey struct{}
type responseWithLogger func(next http.Handler) http.Handler

func (responseWithLogger) WithContex(ctx context.Context, logger *zerolog.Logger) context.Context {
	return context.WithValue(ctx, responseWithLoggerKey{}, logger)
}

var nop = zerolog.Nop()

func (responseWithLogger) FromContext(ctx context.Context) *zerolog.Logger {
	if l, ok := ctx.Value(responseWithLoggerKey{}).(*zerolog.Logger); ok {
		return l
	} else {
		return &nop
	}
}

var ResponseWithLogger responseWithLogger = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger, end := New(r.Context(), log.Logger, w, os.Stderr)
		r = r.WithContext((responseWithLogger)(nil).WithContex(r.Context(), &logger))
		next.ServeHTTP(w, r)
		end()
	})
}
