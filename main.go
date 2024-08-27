package main

import (
	"os"
	"runtime/pprof"

	"github.com/ross96D/updater/cmd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func main() {
	for _, arg := range os.Args {
		if arg == "--profile" {
			file, err := os.Create("cpu_profiler")
			if err != nil {
				panic(err)
			}

			if err = pprof.StartCPUProfile(file); err != nil {
				panic(err)
			}
			defer pprof.StopCPUProfile()
			break
		}
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006-01-02T15:04:05.999999"})

	cobra.EnableTraverseRunHooks = true
	cmd.Execute()
}
