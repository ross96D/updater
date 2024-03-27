package main

import (
	"os"
	"runtime/pprof"

	"github.com/ross96D/updater/cmd"
	"github.com/spf13/cobra"
)

func main() {
	file, err := os.Create("cpu_profiler")
	if err != nil {
		panic(err)
	}

	if err = pprof.StartCPUProfile(file); err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()

	cobra.EnableTraverseRunHooks = true
	cmd.Execute()
}
