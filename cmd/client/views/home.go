package views

import (
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
	Servers     []models.Server
	list        components.List[*models.Server]
	initialized bool
}

func (HomeView) Init() tea.Cmd {
	return homeViewInitializeMsg
}

func (hv HomeView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	}

	if !hv.initialized {
		return hv, Repeat(msg)
	}
	model, cmd := hv.list.Update(msg)
	hv.list = model.(components.List[*models.Server])
	return hv, cmd
}

func (hv HomeView) View() string {
	return hv.list.View()
}

func (hv *HomeView) init() {
	length := len(hv.Servers)
	items := make([]components.Item[*models.Server], 0, length)

	for i := 0; i < length; i++ {
		server := &hv.Servers[i]
		items = append(items, components.Item[*models.Server]{
			Message: server.Name + " " + server.IP,
			Value:   server,
		})
	}

	listModel := components.NewList(items, "", &components.DelegatesKeyMap{
		Select: components.KeyMap{
			Key: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select the file to donwload"),
			),
			Action: func() tea.Cmd {
				return homeViewSelectItemMsg
			},
		},
	})
	hv.list = listModel
	hv.initialized = true
}
