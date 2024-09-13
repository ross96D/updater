package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/models"
	"github.com/ross96D/updater/cmd/client/state"
	"github.com/ross96D/updater/cmd/client/views"
	"github.com/ross96D/updater/server/user_handler"
)

func main() {
	servers := []models.Server{
		{
			ServerName: "server1",
			Url:        models.UnsafeNewURL("190.168.0.1"),
			Apps: []user_handler.App{
				{
					Index: 1,
				},
				{
					Index: 2,
				},
			},
		},
		{
			ServerName: "server2",
			Url:        models.UnsafeNewURL("190.68.0.2"),
			Apps: []user_handler.App{
				{
					Index: 1,
				},
				{
					Index: 2,
				},
			},
		},
	}
	if _, err := tea.NewProgram(views.HomeView{State: state.NewState(servers)}).Run(); err != nil {
		panic(err)
	}
}
