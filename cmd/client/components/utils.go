package components

import tea "github.com/charmbracelet/bubbletea"

func MsgCmd(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

func Repeat(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
