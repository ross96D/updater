package views

import tea "github.com/charmbracelet/bubbletea"

func Repeat(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
