package match_test

import (
	"os"
	"testing"

	"github.com/ross96D/updater/share/configuration"
	"github.com/ross96D/updater/share/match"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

func TestCommand(t *testing.T) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006-01-02T15:04:05.999999"})

	err := match.RunCommand(&log.Logger, configuration.Command{
		Command: "ls",
		Path:    "/",
	})
	require.NoError(t, err)
}
