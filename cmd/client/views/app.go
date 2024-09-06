package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/components"
	"github.com/ross96D/updater/cmd/client/models"
)

type app struct {
	// TODO change []models.Server to a global and easy to access state
	servers   []models.Server
	navigator *components.Navigator
	initCmd   tea.Cmd
}

func NewApp(servers []models.Server) tea.Model {
	nav := new(components.Navigator)
	_, cmd := nav.Push(HomeView{Servers: &servers})
	return &app{
		navigator: nav,
		servers:   servers,
		initCmd:   cmd,
	}
}

func (model *app) Init() tea.Cmd {
	return model.initCmd
}

func (model *app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(InsertServerMsg); ok {
		model.servers = append(model.servers, models.Server(msg))
		return model, nil
	}
	return model, model.navigator.Update(msg)
}

func (model *app) View() string {
	return model.navigator.View()
}
