package streamviewport

import (
	"io"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/share/utils"
)

type endRead struct{ error }

func New(reader io.ReadCloser, width, height int) *Model {
	return &Model{
		viewport: viewport.New(width, height),
		data:     &utils.StreamBuffer{},
		reader:   reader,
	}
}

type Model struct {
	viewport viewport.Model
	data     *utils.StreamBuffer
	reader   io.ReadCloser
	end      bool
}

type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(33*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m *Model) Init() tea.Cmd {
	readCmd := func() tea.Msg {
		_, err := io.Copy(m.data, m.reader)
		m.reader.Close() //nolint:errcheck
		return endRead{err}
	}
	return tea.Batch(doTick(), readCmd)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		m.Sync()
		if m.end {
			return m, nil
		}
		return m, doTick()

	case endRead:
		m.end = true
		// todo set end

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "Q", tea.KeyCtrlC.String():
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	return m.viewport.View()
}

func (m *Model) Sync() tea.Cmd {
	atBottom := m.viewport.AtBottom()
	m.viewport.SetContent(string(m.data.Bytes()))
	if atBottom {
		m.viewport.GotoBottom()
	}
	return viewport.Sync(m.viewport)
}
