package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/components"
	"github.com/ross96D/updater/cmd/client/models"
	"github.com/ross96D/updater/cmd/client/state"
)

type app struct {
	// TODO change []models.Server to a global and easy to access state
	state     *state.GlobalState
	navigator *components.Navigator
	initCmd   tea.Cmd
}

func NewApp(servers []models.Server) tea.Model {
	nav := new(components.Navigator)
	state := state.NewState(servers)
	_, cmd := nav.Push(HomeView{Servers: state})
	return &app{
		navigator: nav,
		state:     state,
		initCmd:   cmd,
	}
}

func (model *app) Init() tea.Cmd {
	return model.initCmd
}

func (model *app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(InsertServerMsg); ok {
		model.state.Add(models.Server(msg))
		return model, state.GlobalStateSyncCmd
	}
	if msg, ok := msg.(EditServerMsg); ok {
		model.state.Set(msg.index, msg.server)
		return model, state.GlobalStateSyncCmd
	}
	return model, model.navigator.Update(msg)
}

func (model *app) View() string {
	return model.navigator.View()
}
