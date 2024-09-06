package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/models"
	"github.com/ross96D/updater/cmd/client/views"
	"github.com/ross96D/updater/server/user_handler"
	"github.com/ross96D/updater/share/configuration"
)

func main() {
	server := models.Server{
		ServerName: "Server1",
		Url:        models.UnsafeNewURL("192.168.0.1"),
		Apps: []user_handler.App{
			{
				Index: 1,
				Application: configuration.Application{
					AuthToken: "token",
					Assets: []configuration.Asset{
						{
							Name:        "Asset1",
							SystemPath:  "path/to",
							ServicePath: "service",
							Unzip:       true,
						},
					},
				},
			},
			{
				Index: 2,
				Application: configuration.Application{
					AuthToken: "token",
					Assets: []configuration.Asset{
						{
							Name:        "Asset1",
							SystemPath:  "path/to",
							ServicePath: "service",
							Unzip:       true,
						},
					},
				},
			},
			{
				Index: 3,
				Application: configuration.Application{
					AuthToken: "token",
					Assets: []configuration.Asset{
						{
							Name:        "Asset1",
							SystemPath:  "path/to",
							ServicePath: "service",
							Unzip:       true,
						},
					},
				},
			},
		},
	}

	var model tea.Model = views.ServerView{Server: server}
	if _, err := tea.NewProgram(model).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
