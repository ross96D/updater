package streamviewport

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ross96D/updater/share/utils"
)

type endRead struct{ error }

type Opt func(m *Model)

func WithTitle(title string) Opt {
	return func(m *Model) {
		m.title = title
	}
}

func New(reader io.ReadCloser, opts ...Opt) *Model {
	m := &Model{
		viewport: viewport.New(0, 0),
		data:     &utils.StreamBuffer{},
		reader:   reader,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

type Model struct {
	viewport viewport.Model
	data     *utils.StreamBuffer
	reader   io.ReadCloser
	title    string
	end      bool
}

type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Batch(
		tea.WindowSize(),
		tea.Tick(33*time.Millisecond, func(t time.Time) tea.Msg {
			return TickMsg(t)
		}),
	)
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

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMarginHeight
		return m, nil

	case endRead:
		m.end = true

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
	return m.headerView() + "\n" + m.viewport.View() + "\n" + m.footerView()
}

func (m *Model) Sync() tea.Cmd {
	atBottom := m.viewport.AtBottom()
	m.viewport.SetContent(string(m.data.Bytes()))
	if atBottom {
		m.viewport.GotoBottom()
	}
	return viewport.Sync(m.viewport)
}

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

func (m *Model) headerView() string {
	var title string
	if m.title != "" {
		title = titleStyle.Render(m.title)
	}
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m *Model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}
