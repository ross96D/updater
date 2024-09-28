package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ross96D/go-utils/list"
	"github.com/ross96D/updater/cmd/client/components"
	listcomp "github.com/ross96D/updater/cmd/client/components/list"
	"github.com/ross96D/updater/cmd/client/components/toast"
	"github.com/ross96D/updater/cmd/client/models"
	"github.com/ross96D/updater/cmd/client/pretty"
	"github.com/ross96D/updater/cmd/client/state"
)

type app struct {
	// TODO change []models.Server to a global and easy to access state
	state         *state.GlobalState
	navigator     *components.Navigator
	initCmd       tea.Cmd
	notifications *list.List[toast.Toast]
	windowSize    tea.WindowSizeMsg
}

func NewApp(state *state.GlobalState) tea.Model {
	var notifications list.List[toast.Toast]
	nav := new(components.Navigator)
	_, cmd := nav.Push(HomeView{State: state})
	return &app{
		navigator:     nav,
		state:         state,
		initCmd:       cmd,
		notifications: &notifications,
	}
}

func (model *app) Init() tea.Cmd {
	return tea.Batch(model.state.FetchCmd(), model.initCmd, tea.WindowSize())
}

func (model *app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.Cmd:
		return model, msg

	case InsertServerMsg:
		model.state.Add(models.Server(msg))
		return model, tea.Batch(state.GlobalStateSyncCmd, state.SaveCmd)

	case EditServerMsg:
		model.state.Set(msg.index, msg.server)
		return model, tea.Batch(state.GlobalStateSyncCmd, state.SaveCmd)

	case state.FetchResultMsg:
		pretty.Print("fetchResult", msg.ServerName, msg.Server.Version.String())
		index := model.state.Find(
			func(s *models.Server) bool {
				return s.ServerName == msg.ServerName
			},
		)
		if index == -1 {
			return model, nil
		}
		server := model.state.Get(index)
		server.Apps = msg.Server.Apps
		server.Version = msg.Server.Version
		server.Status = listcomp.Ready
		model.state.Set(index, server)
		return model, tea.Batch(state.GlobalStateSyncCmd, state.SaveCmd)

	case state.ErrFetchFailMsg:
		t := toast.New(msg.Err.Error(), toast.WithType(toast.Error))
		model.notifications.PushBack(t)

		index := model.state.Find(
			func(s *models.Server) bool {
				return s.ServerName == msg.ServerName
			},
		)
		if index == -1 {
			return model, t.Init()
		}
		server := model.state.Get(index)
		server.Status = listcomp.Error
		model.state.Set(index, server)
		return model, tea.Batch(state.GlobalStateSyncCmd, t.Init())

	case toast.RemoveToastMsg:
		model.removeToast(msg)
		return model, tea.WindowSize()

	case toast.AddToastMsg:
		t := toast.Toast(msg)
		model.notifications.PushBack(toast.Toast(msg))
		return model, t.Init()

	case tea.WindowSizeMsg:
		model.windowSize = msg
		if width := msg.Width - lipgloss.Width(model.notificationsView()); width > 20 {
			msg.Width = width
		}
		return model, model.navigator.Update(msg)
	}

	return model, model.navigator.Update(msg)
}

func (model *app) View() string {
	notificationsView := model.notificationsView()
	width, height := lipgloss.Size(notificationsView)
	notificationsView = lipgloss.Place(width, height, lipgloss.Right, lipgloss.Top, notificationsView)

	return lipgloss.JoinHorizontal(lipgloss.Left, model.navigator.View(), notificationsView)
}

func (model *app) notificationsView() string {
	builder := strings.Builder{}
	for node := range model.notifications.Each {
		builder.WriteString(node.Value.View())
		builder.WriteByte('\n')
	}
	return builder.String()
}

func (model *app) removeToast(msg toast.RemoveToastMsg) {
	for node := range model.notifications.Each {
		if node.Value.ID() == msg.ID {
			model.notifications.Remove(&node)
			break
		}
	}
}
