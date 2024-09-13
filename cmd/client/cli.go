package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/state"
	"github.com/ross96D/updater/cmd/client/views"
	"github.com/spf13/cobra"
)

var cliCommand = &cobra.Command{
	Use: "cli",
	Run: func(cmd *cobra.Command, args []string) {
		state.LoadConfig()
		state := &state.Configuration().State
		program := views.NewApp(state)
		if _, err := tea.NewProgram(program).Run(); err != nil {
			print(err.Error())
			os.Exit(1)
		}
	},
}
