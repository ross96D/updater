package views

import (
	"fmt"
	"io"
	"net/url"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/ross96D/updater/cmd/client/api"
	"github.com/ross96D/updater/cmd/client/components"
	"github.com/ross96D/updater/cmd/client/components/confirmation_dialog"
	"github.com/ross96D/updater/cmd/client/components/list"
	"github.com/ross96D/updater/cmd/client/components/streamviewport"
	"github.com/ross96D/updater/cmd/client/models"
	"github.com/ross96D/updater/cmd/client/state"
	"github.com/ross96D/updater/server/user_handler"
)

type serverViewInitializeMsg struct{}
type serverViewSelectItemMsg struct{}
type serverViewAskUpgradeSelectedMsg struct{}
type serverViewUpgradeSelectedMsg struct{}
type serverViewDryRunUpgradeSelectedMsg struct{}
type serverViewStartStreamPagerMsg struct{ io.ReadCloser }

var serverViewInitializeCmd = func() tea.Msg { return serverViewInitializeMsg{} }
var serverViewSelectItemCmd = func() tea.Msg { return serverViewSelectItemMsg{} }
var serverViewAskUpgradeSelectedCmd = func() tea.Msg { return serverViewAskUpgradeSelectedMsg{} }
var serverViewUpgradeSelectedCmd = func() tea.Msg { return serverViewUpgradeSelectedMsg{} }
var serverViewDryRunUpgradeSelectedCmd = func() tea.Msg { return serverViewDryRunUpgradeSelectedMsg{} }
var serverViewStartStreamPagerCmd = func(rc io.ReadCloser) tea.Cmd {
	return func() tea.Msg {
		return serverViewStartStreamPagerMsg{rc}
	}
}

type ServerView struct {
	Server      models.Server
	list        list.List[*user_handler.App]
	initialized bool
}

func (ServerView) Init() tea.Cmd {
	return serverViewInitializeCmd
}

func (sv ServerView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyCtrlC.String(), "q":
			return sv, tea.Quit
		}

	case serverViewInitializeMsg:
		sv.init()
		return sv, tea.WindowSize()

	case serverViewSelectItemMsg:
		item, ok := sv.list.Selected()
		if !ok {
			return sv, nil
		}
		return sv, components.NavigatorPush(AppView{App: *item.Value})

	case serverViewAskUpgradeSelectedMsg:
		item, ok := sv.list.Selected()
		if !ok {
			return sv, nil
		}
		return sv, components.NavigatorPush(&confirmation_dialog.Model{
			Descripion: fmt.Sprintf("updgrade %s?", item.Value.Name),
			Task:       serverViewUpgradeSelectedCmd,
		})

	case serverViewUpgradeSelectedMsg:
		item, ok := sv.list.Selected()
		if !ok {
			return sv, nil
		}
		return sv, func() tea.Msg {
			session, err := api.NewSession(sv.Server)
			if err != nil {
				return state.ErrFetchFailMsg{ServerName: item.Value.Name, Err: err}
			}
			resp, err := session.Update(*item.Value, false)
			if err != nil {
				return err
			}
			return serverViewStartStreamPagerCmd(resp)
		}

	case serverViewDryRunUpgradeSelectedMsg:
		item, ok := sv.list.Selected()
		if !ok {
			return sv, nil
		}
		return sv, func() tea.Msg {
			session, err := api.NewSession(sv.Server)
			if err != nil {
				return state.ErrFetchFailMsg{ServerName: item.Value.Name, Err: err}
			}
			resp, err := session.Update(*item.Value, true)
			if err != nil {
				return err
			}
			return serverViewStartStreamPagerCmd(resp)
		}

	case serverViewStartStreamPagerMsg:
		return sv, components.NavigatorPush(streamviewport.New(msg))

	case state.GlobalStateSyncMsg:
		sv.init()
		cmd = tea.WindowSize()
	}
	if !sv.initialized {
		return sv, Repeat(msg)
	}

	m, cmd2 := sv.list.Update(msg)
	cmd = tea.Batch(cmd, cmd2)
	sv.list = m.(list.List[*user_handler.App])
	return sv, cmd
}

func (sv ServerView) View() string {
	return sv.list.View()
}

func (sv *ServerView) init() {
	length := len(sv.Server.Apps)
	items := make([]list.Item[*user_handler.App], 0, length)

	for i := 0; i < length; i++ {
		items = append(items, list.Item[*user_handler.App]{
			Message: strconv.Itoa(sv.Server.Apps[i].Index) + " " + ansi.Wrap(sv.Server.Apps[i].Name, 25, "..."),
			Value:   &sv.Server.Apps[i],
		})
	}
	title := sv.Server.ServerName + " IP: " + (*url.URL)(sv.Server.Url).String()
	listModel := list.NewList(items, title, &list.DelegatesKeyMap{
		Select: list.KeyMap{
			Key: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select app"),
			),
			Action: func() tea.Cmd {
				return serverViewSelectItemCmd
			},
		},
		Others: []list.KeyMap{
			{
				Key: key.NewBinding(
					key.WithKeys("u", "U"),
					key.WithHelp("u", "update application"),
				),
				Action: func() tea.Cmd {
					return serverViewAskUpgradeSelectedCmd
				},
			},
			{
				Key: key.NewBinding(
					key.WithKeys("ctrl+u", "ctrl+U"),
					key.WithHelp("ctrl+u", "dry-run update application"),
				),
				Action: func() tea.Cmd {
					return serverViewDryRunUpgradeSelectedCmd
				},
			},
		},
	}, false)

	sv.list = listModel
	sv.initialized = true
}
