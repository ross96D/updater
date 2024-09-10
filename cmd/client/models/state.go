package models

import tea "github.com/charmbracelet/bubbletea"

type GlobalStateSyncMsg struct{}

var GlobalStateSyncCmd = func() tea.Msg {
	return GlobalStateSyncMsg{}
}

type GlobalState struct {
	servers *[]Server
}

func (gs *GlobalState) Len() int {
	return len(*gs.servers)
}

func (gs *GlobalState) Add(s Server) {
	*gs.servers = append((*gs.servers), s)
}

func (gs *GlobalState) Get(i int) Server {
	return (*gs.servers)[i]
}

func (gs *GlobalState) Set(i int, s Server) {
	(*gs.servers)[i] = s
}

func (gs *GlobalState) GetRef(i int) *Server {
	return &(*gs.servers)[i]
}

func NewState(servers []Server) *GlobalState {
	return &GlobalState{
		servers: &servers,
	}
}
