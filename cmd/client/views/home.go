package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/models"
)

type home struct {
	servers []models.Server
}

func NewApp(servers []models.Server) tea.Model {
	return home{
		servers: servers,
	}
}

func (home) Init() tea.Cmd {
	return nil
}

func (model home) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return model, nil
}

func (model home) View() string {
	return ""
}
