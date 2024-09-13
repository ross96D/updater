package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func ConvertToItems[T any](items []Item[T]) []list.Item {
	var interfaceItems []list.Item
	for _, item := range items {
		interfaceItems = append(interfaceItems, item)
	}
	return interfaceItems
}

// TODO Make this view a bit more generic to be used as a component

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
		// Foreground(views.ColorWhite).
		// Background(views.ColorGreen).
		Padding(0, 1)
)

type Item[T any] struct {
	Message string
	Value   T
}

func (i Item[T]) FilterValue() string { return i.Message }
func (i Item[T]) Title() string       { return i.Message }
func (i Item[T]) Description() string { return "" }

// default keys actions
type listKeyMap struct {
	toggleSpinner    key.Binding
	toggleTitleBar   key.Binding
	toggleStatusBar  key.Binding
	togglePagination key.Binding
	toggleHelpMenu   key.Binding
}

type KeyMap struct {
	Key    key.Binding
	Action func() tea.Cmd
}

// custom keys actions
type DelegatesKeyMap struct {
	Select KeyMap
	Others []KeyMap
}

type List[T any] struct {
	list         list.Model
	keys         *listKeyMap
	delegateKeys *DelegatesKeyMap
	quitting     bool

	items []Item[T]
}

type ListOpts[T any] func(*List[T])

func SetListTitle[T any](title string, style lipgloss.Style) ListOpts[T] {
	return func(l *List[T]) {
		l.list.Title = title
		l.list.Styles.Title = style
	}
}

func SetListRawDelegate[T any](delegate list.ItemDelegate) ListOpts[T] {
	return func(l *List[T]) {
		l.list.SetDelegate(delegate)
	}
}

func SetListDelegate[T any](delegateKeys *DelegatesKeyMap, listKeys []key.Binding) ListOpts[T] {
	return func(l *List[T]) {
		delegate := newItemDelegate[T](delegateKeys)
		l.list.SetDelegate(delegate)

		if listKeys != nil {
			l.list.AdditionalFullHelpKeys = func() []key.Binding {
				return listKeys
			}
		}
	}
}

func NewList[T any](items []Item[T], title string, delegateKeys *DelegatesKeyMap, opts ...ListOpts[T]) List[T] {
	var listKeys = newListKeyMap()

	// how to render each item
	delegate := newItemDelegate[T](delegateKeys)
	delegate.ShowDescription = false

	list := list.New(ConvertToItems(items), delegate, 0, 0)
	list.Title = title
	list.Styles.Title = titleStyle
	l := List[T]{
		list:         list,
		keys:         listKeys,
		delegateKeys: delegateKeys,
		items:        items,
	}

	//! rewrite to user options
	for _, v := range opts {
		v(&l)
	}

	return l
}

func (l *List[T]) Selected() Item[T] {
	return l.list.SelectedItem().(Item[T])
}

func (l *List[T]) SelectedIndex() int {
	return l.list.Index()
}

func (l *List[T]) SelectedMessage() (string, error) {
	if l.quitting {
		if item, ok := l.list.SelectedItem().(Item[T]); ok {
			return item.Message, nil
		} else {
			return "", fmt.Errorf("no element selected")
		}
	}
	return l.list.SelectedItem().(Item[T]).Message, fmt.Errorf("no element selected")
}

func (m List[T]) Init() tea.Cmd {
	return nil
}

func (m List[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model, cmd := m.list.Update(msg)
		m.list = model
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		return m, cmd

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.toggleSpinner):
			cmd := m.list.ToggleSpinner()
			return m, cmd

		case key.Matches(msg, m.keys.toggleTitleBar):
			v := !m.list.ShowTitle()
			m.list.SetShowTitle(v)
			m.list.SetShowFilter(v)
			m.list.SetFilteringEnabled(v)
			return m, nil

		case key.Matches(msg, m.keys.toggleStatusBar):
			m.list.SetShowStatusBar(!m.list.ShowStatusBar())
			return m, nil

		case key.Matches(msg, m.keys.togglePagination):
			m.list.SetShowPagination(!m.list.ShowPagination())
			return m, nil

		case key.Matches(msg, m.keys.toggleHelpMenu):
			m.list.SetShowHelp(!m.list.ShowHelp())
			return m, nil

		case key.Matches(msg, m.delegateKeys.Select.Key):
			m.quitting = true
		}
	}

	// This will also call our delegate's update function.
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m List[T]) View() string {
	return appStyle.Render(m.list.View())
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{}
}

func newItemDelegate[T any](keys *DelegatesKeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		if _, ok := m.SelectedItem().(Item[T]); !ok {
			return nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.Select.Key):
				return keys.Select.Action()
			default:
				for _, v := range keys.Others {
					if key.Matches(msg, v.Key) {
						return v.Action()
					}
				}
			}
		}

		return nil
	}

	help := []key.Binding{keys.Select.Key}
	for _, v := range keys.Others {
		help = append(help, v.Key)
	}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}
