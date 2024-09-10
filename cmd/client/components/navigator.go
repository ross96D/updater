package components

import (
	"sync"
	"unsafe"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/models"
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

func NavigatorPush(m tea.Model) tea.Cmd {
	return func() tea.Msg {
		return navigatorPush(m)
	}
}

func NavigatorPop() tea.Msg {
	return navigatorPop{}
}

type AlertPop struct{}

var StopPop = func() tea.Msg { return nil }

type Navigator struct {
	s stack
}

func (nav *Navigator) Push(m tea.Model) (tea.Model, tea.Cmd) {
	nav.s.Push(m)
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
		nav.Pop()
		return models.GlobalStateSyncCmd

	case navigatorPush:
		_, cmd := nav.Push(msg)
		return tea.Sequence(cmd, models.GlobalStateSyncCmd)

	case tea.KeyMsg:
		// navigation go back
		if msg.Type == tea.KeyEscape {
			m, cmd := nav.s.Last().Update(AlertPop{})
			nav.s.SetLast(m)
			if nav.stopBackNavigation(cmd) {
				m, cmd := nav.s.Last().Update(AlertPop{})
				nav.s.SetLast(m)
				return cmd
			}
			return NavigatorPop
		}
	}

	m, cmd := nav.s.Last().Update(msg)
	nav.s.SetLast(m)
	return cmd
}

func (nav *Navigator) stopBackNavigation(cmd tea.Cmd) bool {
	return unsafe.Pointer(&cmd) == unsafe.Pointer(&StopPop)
}
