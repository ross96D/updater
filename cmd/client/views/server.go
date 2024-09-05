package views

import (
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/components"
	"github.com/ross96D/updater/cmd/client/models"
	"github.com/ross96D/updater/server/user_handler"
)

type serverViewInitialize struct{}

type serverViewSelectItem struct{}

var serverViewInitializeMsg = func() tea.Msg { return serverViewInitialize{} }

var serverViewSelectItemMsg = func() tea.Msg { return serverViewSelectItem{} }

type ServerView struct {
	Server      models.Server
	list        components.List[*user_handler.App]
	initialized bool
}

func (ServerView) Init() tea.Cmd {
	return serverViewInitializeMsg
}

func (sv ServerView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyCtrlC.String(), "q":
			return sv, tea.Quit
		}
	case serverViewInitialize:
		sv.init()
		return sv, tea.WindowSize()

	case serverViewSelectItem:
		item, err := sv.list.Selected()
		if err != nil {
			panic(err)
		}
		return sv, components.NavigatorPush(AppView{App: *item.Value})
	}
	if !sv.initialized {
		return sv, Repeat(msg)
	}

	m, cmd := sv.list.Update(msg)
	sv.list = m.(components.List[*user_handler.App])
	return sv, cmd
}

func (sv ServerView) View() string {
	return sv.list.View()
}

func (sv *ServerView) init() {
	length := len(sv.Server.Apps)
	items := make([]components.Item[*user_handler.App], 0, length)

	for i := 0; i < length; i++ {
		items = append(items, components.Item[*user_handler.App]{
			Message: strconv.Itoa(sv.Server.Apps[i].Index) + " TODO missing name",
			Value:   &sv.Server.Apps[i],
		})
	}
	title := sv.Server.Name + " IP: " + sv.Server.IP
	listModel := components.NewList(items, title, &components.DelegatesKeyMap{
		Select: components.KeyMap{
			Key: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select the file to donwload"),
			),
			Action: func() tea.Cmd {
				return serverViewSelectItemMsg
			},
		},
	})

	sv.list = listModel
	sv.initialized = true
}
