package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/views"
	"github.com/ross96D/updater/share/configuration"
)

func main() {
	app := configuration.Application{
		AuthToken: "token",
		Assets: []configuration.Asset{
			{
				Name:        "Asset1",
				SystemPath:  "path/to",
				ServicePath: "service",
				Unzip:       true,
				Command: &configuration.Command{
					Command: "npm",
					Args:    []string{"install", "--omit-dev"},
					Path:    "/pat/to/working/directory",
				},
			},
		},
	}

	if _, err := tea.NewProgram(views.AppView{App: app}).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
