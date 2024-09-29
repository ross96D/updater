package components

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/pretty"
	"github.com/ross96D/updater/cmd/client/state"
)

type stack struct {
	list []tea.Model
	mut  sync.Mutex
}

func (s *stack) ensure() {
	if s.list == nil {
		s.list = make([]tea.Model, 0)
	}
}

func (s *stack) Last() tea.Model {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.ensure()
	return s.list[len(s.list)-1]
}

func (s *stack) SetLast(m tea.Model) {
	s.mut.Lock()
	s.ensure()
	s.list[len(s.list)-1] = m
	s.mut.Unlock()
}

func (s *stack) Push(v tea.Model) {
	s.mut.Lock()
	s.ensure()
	s.list = append(s.list, v)
	s.mut.Unlock()
}

func (s *stack) Pop() tea.Model {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.ensure()

	if len(s.list) == 0 {
		panic("pop on empty stack")
	}

	l := len(s.list)
	if l == 1 {
		return s.list[l-1]
	}
	s.list = s.list[:l-1]

	return s.list[l-2]
}

type navigatorPush tea.Model
type navigatorPop struct{}
type escHandlerMsg bool

func NavigatorPush(m tea.Model) tea.Cmd {
	return func() tea.Msg {
		return navigatorPush(m)
	}
}
func NavigatorPop() tea.Msg {
	return navigatorPop{}
}
func EscHandler(h bool) tea.Cmd {
	return func() tea.Msg {
		return escHandlerMsg(h)
	}
}

type NavModel interface {
	tea.Model
	Enter() tea.Cmd
	Out() tea.Cmd
}

type Navigator struct {
	s         stack
	handleEsc bool
}

func NewNavigator() *Navigator {
	return &Navigator{handleEsc: true}
}

func (nav *Navigator) Push(m tea.Model) (tea.Model, tea.Cmd) {
	nav.s.Push(m)
	if m, ok := m.(NavModel); ok {
		// TODO define if Init or Enter should be called.. probably both should be called? or not
		// is a bit confusing having both here
		return m, tea.Sequence(m.Init(), m.Enter())
	}
	return m, m.Init()
}

func (nav *Navigator) Pop() tea.Model {
	return nav.s.Pop()
}

func (nav *Navigator) View() string {
	return nav.s.Last().View()
}

func (nav *Navigator) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case navigatorPop:
		var cmd [2]tea.Cmd = [2]tea.Cmd{nil, nil}
		if m, ok := nav.Pop().(NavModel); ok {
			cmd[0] = m.Out()
		}
		if m, ok := nav.s.Last().(NavModel); ok {
			cmd[1] = m.Enter()
		}
		return tea.Batch(cmd[0], cmd[1])

	case navigatorPush:
		_, cmd := nav.Push(msg)
		return tea.Sequence(cmd, state.GlobalStateSyncCmd)

	case escHandlerMsg:
		pretty.Print("esc handler msg", msg)
		nav.handleEsc = bool(msg)
		return nil

	case tea.KeyMsg:
		// navigation go back
		if msg.Type == tea.KeyEscape && nav.handleEsc {
			return NavigatorPop
		}
	}

	m, cmd := nav.s.Last().Update(msg)
	nav.s.SetLast(m)
	return cmd
}
