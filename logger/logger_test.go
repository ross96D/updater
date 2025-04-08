package logger_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ross96D/updater/logger"
	"github.com/ross96D/updater/share/utils"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	var path string
	logg := zerolog.New(nil).With().Timestamp().Int("test", 0).Logger()
	testLogger := func(w http.ResponseWriter, r *http.Request) {
		handler, logger, err := logger.New(
			logg,
			r,
			w,
		)
		require.NoError(t, err)
		for i := 0; i < 5000; i++ {
			logger.Info().Msg("test")
		}
		path = filepath.Join(utils.TempDirectory(), handler.FileName())
		handler.End()
	}
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	testLogger(w, req)

	res := w.Result()
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	buff, err := os.ReadFile(path)
	require.NoError(t, err)

	require.Equal(t, string(buff), string(data))
}
