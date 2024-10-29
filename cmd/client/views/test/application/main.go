package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/views"
	"github.com/ross96D/updater/server/user_handler"
	"github.com/ross96D/updater/share/configuration"
)

func main() {
	app := user_handler.App{
		Index: 1,
		Application: configuration.Application{
			AuthToken: "token",
			Assets: []configuration.Asset{
				{
					Name:       "Asset1",
					SystemPath: "path/to",
					Service:    "service",
					Unzip:      true,
					Command: &configuration.Command{
						Command: "npm",
						Args:    []string{"install", "--omit-dev"},
						Path:    "/pat/to/working/directory",
					},
				},
				{
					Name:       "Asset2",
					SystemPath: "path/to",
					Service:    "service",
					Unzip:      true,
					Command: &configuration.Command{
						Command: "npm",
						Args:    []string{"install", "--omit-dev"},
						Path:    "/pat/to/working/directory",
					},
				},
			},
		},
	}

	if _, err := tea.NewProgram(views.AppView{App: app}).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
