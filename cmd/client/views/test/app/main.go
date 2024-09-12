package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/models"
	"github.com/ross96D/updater/cmd/client/pretty"
	"github.com/ross96D/updater/cmd/client/views"
	"github.com/ross96D/updater/server/user_handler"
	"github.com/ross96D/updater/share/configuration"
)

func main() {
	pretty.ActivateDebug()
	servers := []models.Server{
		{
			ServerName: "server1",
			Url:        models.UnsafeNewURL("190.168.0.1"),
			Apps: []user_handler.App{
				{
					Index: 1,
					Application: configuration.Application{
						AuthToken: "token",
						Assets: []configuration.Asset{
							{
								Name:        "asset1",
								ServicePath: "service1",
							},
							{
								Name:        "asset2",
								ServicePath: "service2",
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
								Name:        "asset1",
								ServicePath: "service1",
							},
							{
								Name:        "asset2",
								ServicePath: "service2",
							},
						},
					},
				},
			},
		},
		{
			ServerName: "server2",
			Url:        models.UnsafeNewURL("190.68.0.2"),
			Apps: []user_handler.App{
				{
					Index: 1,
					Application: configuration.Application{
						AuthToken: "token",
						Assets: []configuration.Asset{
							{
								Name:        "asset1",
								ServicePath: "service1",
							},
							{
								Name:        "asset2",
								ServicePath: "service2",
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
								Name:        "asset1",
								ServicePath: "service1",
							},
							{
								Name:        "asset2",
								ServicePath: "service2",
							},
						},
					},
				},
			},
		},
	}
	m := views.NewApp(servers)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		panic(err)
	}
}
