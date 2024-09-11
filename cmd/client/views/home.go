package views

import (
	"net/url"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/components"
	"github.com/ross96D/updater/cmd/client/models"
	"github.com/ross96D/updater/cmd/client/state"
)

type homeViewInitialize struct{}

type homeViewSelectItem struct{}

type homeEditSelectedMsg struct{}

var homeViewInitializeMsg = func() tea.Msg { return homeViewInitialize{} }

var homeViewSelectItemMsg = func() tea.Msg { return homeViewSelectItem{} }

var homeEditSelectedCmd = func() tea.Msg { return homeEditSelectedMsg{} }

type HomeView struct {
	Servers     *state.GlobalState
	list        components.List[*models.Server]
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
		item := hv.list.Selected()
		return hv, components.NavigatorPush(ServerView{Server: *item.Value})

	case homeEditSelectedMsg:
		item := hv.list.Selected()
		index := hv.list.SelectedIndex()
		m := NewServerFormView(&EditServer{server: *item.Value, index: index})
		return hv, components.NavigatorPush(m)

	case state.GlobalStateSyncMsg:
		hv.init()
		cmd = tea.WindowSize()
	}

	if !hv.initialized {
		return hv, Repeat(msg)
	}
	model, cmd2 := hv.list.Update(msg)
	cmd = tea.Batch(cmd, cmd2)
	hv.list = model.(components.List[*models.Server])
	return hv, cmd
}

func (hv HomeView) View() string {
	return hv.list.View()
}

func (hv *HomeView) init() {
	length := hv.Servers.Len()
	items := make([]components.Item[*models.Server], 0, length)

	for i := 0; i < length; i++ {
		server := hv.Servers.GetRef(i)
		items = append(items, components.Item[*models.Server]{
			Message: server.ServerName + " " + (*url.URL)(server.Url).String(),
			Value:   server,
		})
	}

	listModel := components.NewList(items, "",
		&components.DelegatesKeyMap{
			Select: components.KeyMap{
				Key: key.NewBinding(
					key.WithKeys("enter"),
					key.WithHelp("enter", "select server"),
				),
				Action: func() tea.Cmd {
					return homeViewSelectItemMsg
				},
			},
			Others: []components.KeyMap{
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
						key.WithHelp("e", "edit selected server"),
					),
					Action: func() tea.Cmd {
						return homeEditSelectedCmd
					},
				},
			},
		},
	)
	hv.list = listModel
	hv.initialized = true
}
