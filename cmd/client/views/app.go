package views

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ross96D/updater/cmd/client/fpew"
	"github.com/ross96D/updater/share/configuration"
)

type appViewInitialize struct{}

type AppView struct {
	App      configuration.Application
	viewPort *viewport.Model
}

func (av AppView) Init() tea.Cmd {
	return tea.Sequence(tea.WindowSize(), func() tea.Msg { return appViewInitialize{} })
}

func (av AppView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		fpew.Dump("initializing")
		av.viewPort.SetContent(av.content())

	case tea.WindowSizeMsg:
		av.viewPort.Height = msg.Height
		av.viewPort.Width = msg.Width
	}

	return av, nil
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

	for _, asset := range av.App.Assets {
		builder.WriteString(asset.Name + "\n")
		builder.WriteString(keyStyle.Render("service: ") + "\t" + asset.ServicePath + "\n")
		builder.WriteString(keyStyle.Render("system path: ") + "\t" + asset.SystemPath + "\n")
		builder.WriteString(keyStyle.Render("unzip: ") + "\t" + strconv.FormatBool(asset.Unzip) + "\n")
		if asset.Command != nil {
			builder.WriteString(keyStyle.Render("command: ") + "\t" + asset.Command.String() + "\n")
		}
	}
	return builder.String()
}
