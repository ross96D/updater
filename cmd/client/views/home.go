package views

import (
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/api"
	"github.com/ross96D/updater/cmd/client/components"
	"github.com/ross96D/updater/cmd/client/components/confirmation_dialog"
	"github.com/ross96D/updater/cmd/client/components/input_text"
	"github.com/ross96D/updater/cmd/client/components/list"
	"github.com/ross96D/updater/cmd/client/components/toast"
	"github.com/ross96D/updater/cmd/client/models"
	"github.com/ross96D/updater/cmd/client/pretty"
	"github.com/ross96D/updater/cmd/client/state"
)

type homeViewInitialize struct{}
type homeViewSelectItem struct{}
type homeEditSelectedMsg struct{}
type homeAskDeleteSelectedMsg struct{}
type homeDeleteSelectedMsg struct{}
type homeAskUpgradeSelectedMsg struct{}
type homeUpgradeSelectedMsg struct{}
type homeRequestConfigMsg struct{}
type homeConfigSelectedMsg struct{ string }
type homeRequestReloadMsg struct{}
type homeReloadSelectedMsg struct{ string }

var homeViewInitializeMsg = func() tea.Msg { return homeViewInitialize{} }
var homeViewSelectItemMsg = func() tea.Msg { return homeViewSelectItem{} }
var homeEditSelectedCmd = func() tea.Msg { return homeEditSelectedMsg{} }
var homeAskDeleteSelectedCmd = func() tea.Msg { return homeAskDeleteSelectedMsg{} }
var homeDeleteSelectedCmd = func() tea.Msg { return homeDeleteSelectedMsg{} }
var homeAskUpgradeSelectedCmd = func() tea.Msg { return homeAskUpgradeSelectedMsg{} }
var homeUpgradeSelectedCmd = func() tea.Msg { return homeUpgradeSelectedMsg{} }
var homeRequestConfigCmd = func() tea.Msg { return homeRequestConfigMsg{} }
var homeConfigSelectedCmd = func(path string) tea.Cmd { return func() tea.Msg { return homeConfigSelectedMsg{path} } }
var homeRequestReloadCmd = func() tea.Msg { return homeRequestReloadMsg{} }
var homeReloadSelectedCmd = func(path string) tea.Cmd { return func() tea.Msg { return homeReloadSelectedMsg{path} } }

type HomeView struct {
	State       *state.GlobalState
	list        list.List[*models.Server]
	initialized bool
}

func (HomeView) Init() tea.Cmd {
	return homeViewInitializeMsg
}

func (hv HomeView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyCtrlC.String(), "q":
			return hv, tea.Quit
		}
	case homeViewInitialize:
		hv.init()
		return hv, nil

	case homeViewSelectItem:
		item, ok := hv.list.Selected()
		if !ok {
			return hv, nil
		}
		return hv, components.NavigatorPush(ServerView{Server: *item.Value})

	case homeEditSelectedMsg:
		item, ok := hv.list.Selected()
		if !ok {
			return hv, nil
		}
		index := hv.list.SelectedIndex()
		m := NewServerFormView(&EditServer{server: *item.Value, index: index})
		return hv, components.NavigatorPush(m)

	case homeAskDeleteSelectedMsg:
		item, ok := hv.list.Selected()
		if !ok {
			return hv, nil
		}
		return hv, components.NavigatorPush(&confirmation_dialog.Model{
			Descripion: fmt.Sprintf("updagrade %s?", item.Value.ServerName),
			Task:       homeDeleteSelectedCmd,
		})

	case homeDeleteSelectedMsg:
		_, ok := hv.list.Selected()
		if !ok {
			return hv, nil
		}
		index := hv.list.SelectedIndex()
		hv.State.Remove(index)
		return hv, tea.Batch(state.GlobalStateSyncCmd, state.SaveCmd)

	case homeAskUpgradeSelectedMsg:
		item, ok := hv.list.Selected()
		if !ok {
			return hv, nil
		}
		return hv, components.NavigatorPush(&confirmation_dialog.Model{
			Descripion: fmt.Sprintf("updgrade %s?", item.Value.ServerName),
			Task:       homeUpgradeSelectedCmd,
		})

	case homeUpgradeSelectedMsg:
		item, ok := hv.list.Selected()
		if !ok {
			return hv, nil
		}
		upgCmd := func() tea.Msg {
			server := *item.Value
			pretty.Print("getting session ")
			session, err := api.NewSession(server)
			if err != nil {
				pretty.Print("error getting session", err.Error())
				return state.ErrFetchFailMsg{ServerName: server.ServerName, Err: err}
			}
			pretty.Print("getted session, upgrading")
			resp, err := session.Upgrade()
			if err != nil {
				pretty.Print("err upgrading", err.Error(), resp)
				return state.ErrFetchFailMsg{ServerName: server.ServerName, Err: err}
			}
			pretty.Print("upgrade done")
			updateServerCmd := func() tea.Msg {
				pretty.Print("starting udpdate of server data")
				s, err := session.List()
				if err != nil {
					pretty.Print("error on udpdate of server", err.Error())
					return state.ErrFetchFailCmd(server.ServerName, err)
				}
				pretty.Print("udpdate of server data done")
				return state.FetchResultMsg{ServerName: server.ServerName, Server: s}
			}
			return tea.Batch(components.Repeat(toast.AddToastMsg(toast.New(resp))), updateServerCmd)
		}
		addToastCmd := func() tea.Msg {
			return toast.AddToastMsg(toast.New(
				fmt.Sprintf("upgrading server %s", item.Value.ServerName)))
		}
		return hv, tea.Sequence(addToastCmd, upgCmd)

	case homeRequestConfigMsg:
		return hv, components.NavigatorPush(&input_text.Model{
			Title:     "downloaded configuration file name",
			AcceptKey: tea.Key{Type: tea.KeyEnter},
			AcceptCmd: homeConfigSelectedCmd,
		})

	case homeConfigSelectedMsg:
		item, ok := hv.list.Selected()
		if !ok {
			return hv, nil
		}
		configCmd := func() tea.Msg {
			server := *item.Value
			session, err := api.NewSession(server)
			if err != nil {
				return state.ErrFetchFailMsg{ServerName: server.ServerName, Err: err}
			}
			resp, err := session.Config()
			if err != nil {
				return state.ErrFetchFailMsg{ServerName: server.ServerName, Err: err}
			}
			defer resp.Close()
			f, err := os.Create(msg.string)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(f, resp)
			if err != nil {
				return state.ErrFetchFailMsg{ServerName: server.ServerName, Err: fmt.Errorf("io.Copy %w", err)}
			}
			return toast.AddToastMsg(toast.New("config saved to " + msg.string))
		}
		return hv, configCmd

	case homeRequestReloadMsg:
		return hv, components.NavigatorPush(&input_text.Model{
			Title:     "downloaded configuration file name",
			AcceptKey: tea.Key{Type: tea.KeyEnter},
			AcceptCmd: homeReloadSelectedCmd,
		})

	case homeReloadSelectedMsg:
		item, ok := hv.list.Selected()
		if !ok {
			return hv, nil
		}
		reloadCmd := func() tea.Msg {
			server := *item.Value
			session, err := api.NewSession(server)
			if err != nil {
				return state.ErrFetchFailMsg{ServerName: server.ServerName, Err: err}
			}
			f, err := os.Open(msg.string)
			if err != nil {
				return err
			}

			_, err = session.Reload(f)
			if err != nil {
				return state.ErrFetchFailMsg{ServerName: server.ServerName, Err: err}
			}
			return tea.Batch(
				state.Configuration().State.FetchCmdBy(server),
				components.MsgCmd(toast.AddToastMsg(toast.New("config reloaded from "+msg.string))),
			)
		}
		return hv, reloadCmd

	case state.GlobalStateSyncMsg:
		hv.init()
		cmd = tea.WindowSize()
	}

	if !hv.initialized {
		return hv, components.Repeat(msg)
	}
	model, cmd2 := hv.list.Update(msg)
	cmd = tea.Batch(cmd, cmd2)
	hv.list = model.(list.List[*models.Server])
	return hv, cmd
}

func (hv HomeView) View() string {
	return hv.list.View()
}

func (hv *HomeView) init() {
	var selected int
	if hv.initialized {
		selected = hv.list.SelectedIndex()
	}

	length := hv.State.Len()
	if selected >= length {
		selected = length - 1
	}
	if selected < 0 {
		selected = 0
	}

	items := make([]list.Item[*models.Server], 0, length)

	for i := 0; i < length; i++ {
		server := hv.State.GetRef(i)
		items = append(items, list.Item[*models.Server]{
			Message: server.Version.String() + " " +
				server.ServerName + " " +
				(*url.URL)(server.Url).String(),
			Value:       server,
			StatusValue: server.Status,
		})
	}

	listModel := list.NewList(items, "",
		&list.DelegatesKeyMap{
			Select: list.KeyMap{
				Key: key.NewBinding(
					key.WithKeys("enter"),
					key.WithHelp("enter", "select server"),
				),
				Action: func() tea.Cmd {
					return homeViewSelectItemMsg
				},
			},
			Others: []list.KeyMap{
				{
					Key: key.NewBinding(
						key.WithKeys("a", "A"),
						key.WithHelp("a", "add server"),
					),
					Action: func() tea.Cmd {
						return components.NavigatorPush(NewServerFormView(nil))
					},
				},
				{
					Key: key.NewBinding(
						key.WithKeys("e", "E"),
						key.WithHelp("e", "edit server"),
					),
					Action: func() tea.Cmd {
						return homeEditSelectedCmd
					},
				},
				{
					Key: key.NewBinding(
						key.WithKeys("d", "D"),
						key.WithHelp("d", "delete selected server"),
					),
					Action: func() tea.Cmd {
						return homeAskDeleteSelectedCmd
					},
				},
				{
					Key: key.NewBinding(
						key.WithKeys("u", "U"),
						key.WithHelp("u", "updgrade server"),
					),
					Action: func() tea.Cmd {
						return homeAskUpgradeSelectedCmd
					},
				},
				{
					Key: key.NewBinding(
						key.WithKeys("c", "C"),
						key.WithHelp("c", "updgrade server"),
					),
					Action: func() tea.Cmd {
						return homeRequestConfigCmd
					},
				},
				{
					Key: key.NewBinding(
						key.WithKeys("r", "R"),
						key.WithHelp("r", "updgrade server"),
					),
					Action: func() tea.Cmd {
						return homeRequestReloadCmd
					},
				},
			},
		},
		true,
	)
	hv.list = listModel
	hv.initialized = true
	hv.list.SetSelected(selected)
}
