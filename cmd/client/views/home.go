package views

import (
	"fmt"
	"net/url"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/api"
	"github.com/ross96D/updater/cmd/client/components"
	"github.com/ross96D/updater/cmd/client/components/confirmation_dialog"
	"github.com/ross96D/updater/cmd/client/components/list"
	"github.com/ross96D/updater/cmd/client/components/toast"
	"github.com/ross96D/updater/cmd/client/models"
	"github.com/ross96D/updater/cmd/client/state"
)

type homeViewInitialize struct{}
type homeViewSelectItem struct{}
type homeEditSelectedMsg struct{}
type homeAskDeleteSelectedMsg struct{}
type homeDeleteSelectedMsg struct{}
type homeAskUpgradeSelectedMsg struct{}
type homeUpgradeSelectedMsg struct{}

var homeViewInitializeMsg = func() tea.Msg { return homeViewInitialize{} }
var homeViewSelectItemMsg = func() tea.Msg { return homeViewSelectItem{} }
var homeEditSelectedCmd = func() tea.Msg { return homeEditSelectedMsg{} }
var homeAskDeleteSelectedCmd = func() tea.Msg { return homeAskDeleteSelectedMsg{} }
var homeDeleteSelectedCmd = func() tea.Msg { return homeDeleteSelectedMsg{} }
var homeAskUpgradeSelectedCmd = func() tea.Msg { return homeAskUpgradeSelectedMsg{} }
var homeUpgradeSelectedCmd = func() tea.Msg { return homeUpgradeSelectedMsg{} }

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
		return hv, func() tea.Msg {
			session, err := api.NewSession(*item.Value)
			if err != nil {
				return state.ErrFetchFailMsg{ServerName: item.Value.ServerName, Err: err}
			}
			resp, err := session.Upgrade()
			if err != nil {
				return err
			}
			return toast.AddToastMsg(toast.New(resp))
		}

	case state.GlobalStateSyncMsg:
		hv.init()
		cmd = tea.WindowSize()
	}

	if !hv.initialized {
		return hv, Repeat(msg)
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
			Message:     server.ServerName + " " + (*url.URL)(server.Url).String(),
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
						key.WithHelp("e", "edit selected server"),
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
						key.WithHelp("u", "send an updgrade request for the selected updater server"),
					),
					Action: func() tea.Cmd {
						return homeAskUpgradeSelectedCmd
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
