package streamviewport

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ross96D/updater/cmd/client/components"
	"github.com/ross96D/updater/cmd/client/pretty"
	"github.com/ross96D/updater/share/utils"
)

type endRead struct{ error }

type saveFileMsg struct{ filepath string }

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

	input *textinput.Model
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

func (m *Model) Enter() tea.Cmd { return nil }

func (m *Model) Out() tea.Cmd {
	return components.EscHandler(true)
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

	case saveFileMsg:
		path, err := filepath.Abs(msg.filepath)
		if err != nil {
			panic(err)
		}
		return m, func() tea.Msg {
			pretty.Print("creating file", path)
			f, err := os.Create(path)
			if err != nil {
				pretty.Print("creating file err", path, err.Error())
				return err
			}
			_, err = f.Write(m.data.Bytes())
			if err != nil {
				pretty.Print("writing to ", path, err.Error())
			}
			return err
		}

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		if m.input == nil {
			switch msg.String() {
			case "q", "Q", tea.KeyCtrlC.String():
				return m, tea.Quit
			case "S", "s":
				input := textinput.New()
				m.input = &input
				return m, tea.Sequence(components.EscHandler(false), m.input.Focus())
			}
		} else {
			if msg.Type == tea.KeyEnter {
				return m, components.MsgCmd(saveFileMsg{filepath: m.input.Value()})
			} else if msg.Type == tea.KeyEsc {
				m.input = nil
				return m, components.EscHandler(true)
			}
			var cmd tea.Cmd
			*m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
	}
	var cmd [2]tea.Cmd = [2]tea.Cmd{nil, nil}
	m.viewport, cmd[0] = m.viewport.Update(msg)
	if m.input != nil {
		*m.input, cmd[1] = m.input.Update(msg)
	}
	return m, tea.Batch(cmd[0], cmd[1])
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
	if m.input != nil {
		title = "insert filepath: " + m.input.View()
	} else if m.title != "" {
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
