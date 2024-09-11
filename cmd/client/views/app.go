package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/components"
	"github.com/ross96D/updater/cmd/client/models"
	"github.com/ross96D/updater/cmd/client/pretty"
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
	return tea.Batch(model.state.FetchCmd(), model.initCmd)
}

func (model *app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.Cmd:
		return model, msg

	case InsertServerMsg:
		model.state.Add(models.Server(msg))
		return model, state.GlobalStateSyncCmd

	case EditServerMsg:
		model.state.Set(msg.index, msg.server)
		return model, state.GlobalStateSyncCmd

	case state.FetchResultMsg:
		index := model.state.Find(
			func(s *models.Server) bool {
				return s.ServerName == msg.ServerName
			},
		)
		if index == -1 {
			return model, nil
		}
		server := model.state.Get(index)
		server.Apps = msg.Apps
		model.state.Set(index, server)
		return model, state.GlobalStateSyncCmd

	case state.ErrFetchFailMsg:
		// TODO send a toast notification
		pretty.Print(msg, msg.Err.Error())
		return model, nil
	}

	return model, model.navigator.Update(msg)
}

func (model *app) View() string {
	return model.navigator.View()
}
