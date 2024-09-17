package confirmation_dialog

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ross96D/updater/cmd/client/components"
)

type Model struct {
	Descripion  string
	Task        tea.Cmd
	windowsSize tea.WindowSizeMsg
}

func (m *Model) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			return m, tea.Sequence(components.NavigatorPop, m.Task)

		case "n", "N":
			return m, components.NavigatorPop
		}
	case tea.WindowSizeMsg:
		m.windowsSize = tea.WindowSizeMsg{}
	}
	return m, nil
}

func (m *Model) View() string {
	style := lipgloss.NewStyle().
		Width(m.windowsSize.Width).
		Height(m.windowsSize.Height).
		Align(lipgloss.Center)

	return style.Render(m.Descripion + " y/n")
}
