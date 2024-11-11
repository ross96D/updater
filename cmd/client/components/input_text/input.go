package input_text

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/components"
	"github.com/ross96D/updater/cmd/client/pretty"
)

type Model struct {
	Title     string
	Input     textinput.Model
	AcceptKey tea.Key
	AcceptCmd func(string) tea.Cmd
	acceptCmd func(string) tea.Cmd
}

func (m *Model) Init() tea.Cmd {
	if m.AcceptKey.Type == tea.KeyRunes {
		panic("AcceptKey type cannot be of type runes")
	}
	m.acceptCmd = nil
	m.Input = textinput.New()
	return m.Input.Focus()
}

func (m *Model) Enter() tea.Cmd { return nil }
func (m *Model) Out() tea.Cmd {
	if m.acceptCmd != nil {
		return m.acceptCmd(m.Input.Value())
	} else {
		return nil
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case m.AcceptKey.Type:
			m.acceptCmd = m.AcceptCmd
			pretty.Print(m.AcceptCmd == nil, m.acceptCmd == nil)
			return m, components.NavigatorPop

		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyRunes:
			if msg.String() == "q" {
				return m, tea.Quit
			}
		}
	}
	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	return m.Title + "\n" + m.Input.View()
}
