package main

import (
	"github.com/ross96D/updater/cmd/client/state"
	"github.com/spf13/cobra"
)

var cliCommand = &cobra.Command{
	Use: "cli",
	Run: func(cmd *cobra.Command, args []string) {
		state.LoadConfig()
	},
}
