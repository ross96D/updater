package main

import (
	"os"
	"runtime/pprof"

	"github.com/ross96D/updater/cmd"
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

	f, err := os.Create(".log")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	log.Output(f)

	cobra.EnableTraverseRunHooks = true
	cmd.Execute()
}
