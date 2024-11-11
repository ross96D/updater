package state

import (
	"bytes"
	"encoding/json"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/api"
	"github.com/ross96D/updater/cmd/client/models"
	"github.com/ross96D/updater/server/user_handler"
)

type GlobalStateSyncMsg struct{}

var GlobalStateSyncCmd = func() tea.Msg {
	return GlobalStateSyncMsg{}
}

type GlobalState struct {
	servers *[]models.Server
}

func (gs GlobalState) Servers() []models.Server {
	if gs.servers != nil {
		return *gs.servers
	}
	return nil
}

func (gs GlobalState) MarshalJSON() ([]byte, error) {
	buff := bytes.Buffer{}
	if gs.servers == nil {
		buff.WriteString("null")
		return buff.Bytes(), nil
	}
	buff.WriteString("{\"servers\":")
	buff.WriteByte('[')
	for i, v := range *gs.servers {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		buff.Write(b)
		if i != len(*gs.servers)-1 {
			buff.WriteByte(',')
		}
	}
	buff.WriteByte(']')
	buff.WriteByte('}')

	return buff.Bytes(), nil
}

func (gs *GlobalState) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		gs = nil
		return nil
	}

	b = b[bytes.IndexRune(b, ':')+1 : len(b)-1]

	servers := []models.Server{}
	err := json.Unmarshal(b, &servers)
	if err != nil {
		return err
	}
	gs.servers = &servers
	return nil
}

func (gs *GlobalState) Len() int {
	return len(*gs.servers)
}

func (gs *GlobalState) Find(f func(s *models.Server) bool) int {
	for i, server := range *gs.servers {
		if f(&server) {
			return i
		}
	}
	return -1
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

func (gs *GlobalState) Remove(i int) {
	*gs.servers = slices.Delete(*gs.servers, i, i+1)
}

func (gs *GlobalState) GetRef(i int) *models.Server {
	return &(*gs.servers)[i]
}

func NewState(servers []models.Server) *GlobalState {
	return &GlobalState{
		servers: &servers,
	}
}

type FetchResultMsg struct {
	ServerName string
	Server     user_handler.Server
}

type ErrFetchFailMsg struct {
	ServerName string
	Err        error
}

var ErrFetchFailCmd = func(name string, err error) tea.Cmd {
	if err == nil {
		return nil
	}
	return func() tea.Msg {
		return ErrFetchFailMsg{ServerName: name, Err: err}
	}
}

func (gs *GlobalState) FetchCmdBy(server models.Server) tea.Cmd {
	if gs.servers == nil {
		gs.servers = &[]models.Server{}
	}
	f := func(server models.Server) tea.Cmd {
		return func() tea.Msg {
			session, err := api.NewSession(server)
			if err != nil {
				return ErrFetchFailCmd(server.ServerName, err)
			}
			s, err := session.List()
			if err != nil {
				return ErrFetchFailCmd(server.ServerName, err)
			}
			return FetchResultMsg{ServerName: server.ServerName, Server: s}
		}
	}
	for _, s := range *gs.servers {
		if s.ServerName == server.ServerName {
			return f(server)
		}
	}
	return nil
}

func (gs *GlobalState) FetchCmd() tea.Cmd {
	if gs.servers == nil {
		gs.servers = &[]models.Server{}
	}

	cmds := make([]tea.Cmd, 0)

	f := func(server models.Server) tea.Cmd {
		return func() tea.Msg {
			session, err := api.NewSession(server)
			if err != nil {
				return ErrFetchFailCmd(server.ServerName, err)
			}
			s, err := session.List()
			if err != nil {
				return ErrFetchFailCmd(server.ServerName, err)
			}
			return FetchResultMsg{ServerName: server.ServerName, Server: s}
		}
	}
	for _, server := range *gs.servers {
		cmds = append(cmds, f(server))
	}
	return tea.Batch(cmds...)
}
