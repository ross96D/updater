package views

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ross96D/updater/cmd/client/state"
	"github.com/ross96D/updater/server/user_handler"
)

type appViewInitialize struct{}

var appViewInitializeMsg = func() tea.Msg { return appViewInitialize{} }

type AppView struct {
	App      user_handler.App
	viewPort *viewport.Model
}

func (av AppView) Init() tea.Cmd {
	return tea.Sequence(tea.WindowSize(), appViewInitializeMsg)
}

func (av AppView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if av.viewPort == nil {
		v := viewport.New(0, 0)
		av.viewPort = &v
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyCtrlC.String(), "q":
			return av, tea.Quit
		}

	case appViewInitialize:
		av.init()
		return av, tea.WindowSize()

	case tea.WindowSizeMsg:
		av.viewPort.Height = msg.Height - 2
		av.viewPort.Width = msg.Width

	case state.GlobalStateSyncMsg:
		av.init()
		cmd = tea.WindowSize()
	}

	v, cmd2 := av.viewPort.Update(msg)
	cmd = tea.Batch(cmd, cmd2)
	av.viewPort = &v

	return av, cmd
}

func (av AppView) View() string {
	if av.viewPort == nil {
		return ""
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		"TODO: app name is missing",
		av.viewPort.View(),
		"Press u to update the application",
	)
}

func (av AppView) content() string {
	builder := strings.Builder{}

	l := len("system path: ")

	keyStyle := lipgloss.NewStyle().Width(l)

	const ident = "\t"

	for _, asset := range av.App.Assets {
		builder.WriteString(asset.Name + "\n")
		builder.WriteString(ident + keyStyle.Render("service: ") + "\t" + asset.ServicePath + "\n")
		builder.WriteString(ident + keyStyle.Render("system path: ") + "\t" + asset.SystemPath + "\n")
		builder.WriteString(ident + keyStyle.Render("unzip: ") + "\t" + strconv.FormatBool(asset.Unzip) + "\n")
		if asset.Command != nil {
			builder.WriteString(ident + keyStyle.Render("command: ") + "\t" + asset.Command.String() + "\n")
		}
	}
	return builder.String()
}

func (av *AppView) init() {
	av.viewPort.SetContent(av.content())
}
