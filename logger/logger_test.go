package logger_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ross96D/updater/logger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	buff := bytes.Buffer{}
	consoleWriter := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.Out = &buff
	})
	testLogger := func(w http.ResponseWriter, r *http.Request) {
		logger, endFunc := logger.New(
			r.Context(),
			zerolog.New(nil).With().Timestamp().Int("test", 0).Logger(),
			w,
			consoleWriter,
		)
		for i := 0; i < 5000; i++ {
			logger.Info().Msg("test")
		}
		endFunc()
	}
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	testLogger(w, req)

	res := w.Result()
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equal(t, buff.String(), string(data))
}
