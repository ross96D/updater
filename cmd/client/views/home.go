package views

import (
	"net/url"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/components"
	"github.com/ross96D/updater/cmd/client/models"
)

type homeViewInitialize struct{}

type homeViewSelectItem struct{}

var homeViewInitializeMsg = func() tea.Msg { return homeViewInitialize{} }

var homeViewSelectItemMsg = func() tea.Msg { return homeViewSelectItem{} }

type HomeView struct {
	Servers     *models.GlobalState
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
		item, err := hv.list.Selected()
		if err != nil {
			panic(err)
		}
		return hv, components.NavigatorPush(ServerView{Server: *item.Value})
	case models.GlobalStateSyncMsg:
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
						return components.NavigatorPush(NewServerFormView())
					},
				},
			},
		},
	)
	hv.list = listModel
	hv.initialized = true
}
