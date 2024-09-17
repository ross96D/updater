package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/pretty"
	"github.com/ross96D/updater/cmd/client/state"
	"github.com/ross96D/updater/cmd/client/views"
	mod_upgrade "github.com/ross96D/updater/upgrade"
	"github.com/spf13/cobra"
)

var cliDebug bool

var upgradeFlag bool

var cliCommand = &cobra.Command{
	Use: "cli",
	Run: func(cmd *cobra.Command, args []string) {
		if cliDebug {
			// nolint
			pretty.ActivateDebug()
		}

		if upgradeFlag {
			if err := upgradeCli(); err != nil {
				println(err.Error())
				os.Exit(1)
			}
			return
		}

		state.LoadConfig()
		state := &state.Configuration().State
		program := views.NewApp(state)
		if _, err := tea.NewProgram(program).Run(); err != nil {
			print(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	cliCommand.Flags().BoolVar(&cliDebug, "debug", false, "create a debug.log file on the cwd")
	cliCommand.Flags().BoolVar(&upgradeFlag, "upgrade", false, "upgrade to the latest version")
}

func upgradeCli() error {
	return mod_upgrade.Upgrade(mod_upgrade.Client)
}
