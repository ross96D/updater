package state

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/api"
	"github.com/ross96D/updater/cmd/client/models"
)

type GlobalStateSyncMsg struct{}

var GlobalStateSyncCmd = func() tea.Msg {
	return GlobalStateSyncMsg{}
}

type GlobalState struct {
	servers *[]models.Server
}

func (gs *GlobalState) Len() int {
	return len(*gs.servers)
}

func (gs *GlobalState) Add(s models.Server) {
	*gs.servers = append((*gs.servers), s)
}

func (gs *GlobalState) Get(i int) models.Server {
	return (*gs.servers)[i]
}

func (gs *GlobalState) Set(i int, s models.Server) {
	(*gs.servers)[i] = s
}

func (gs *GlobalState) GetRef(i int) *models.Server {
	return &(*gs.servers)[i]
}

func NewState(servers []models.Server) *GlobalState {
	return &GlobalState{
		servers: &servers,
	}
}

func (gs *GlobalState) FetchCmd() tea.Cmd {
	cmds := make([]tea.Cmd, 0)

	f := func(server models.Server) tea.Cmd {
		return func() tea.Msg {
			session, err := api.NewSession(server)
			if err != nil {
				panic(err)
			}
			apps, err := session.List()
			if err != nil {
				panic(err)
			}

			return apps
		}
	}
	for _, server := range *gs.servers {
		cmds = append(cmds, f(server))
	}
	return tea.Batch(cmds...)
}
