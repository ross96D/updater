package toast

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

type RemoveToastMsg struct {
	ID uuid.UUID
}

type AddToastMsg Toast

type toastType int

const (
	Info toastType = iota + 1
	Warn
	Error
)

type ToastOpts = func(*Toast)

func WithDuration(dur time.Duration) ToastOpts {
	return func(t *Toast) {
		t.dur = dur
	}
}
func WithType(toastType toastType) ToastOpts {
	return func(t *Toast) {
		t.toastType = toastType
	}
}

func New(text string, opts ...ToastOpts) Toast {
	toast := Toast{
		text:      text,
		dur:       5 * time.Second,
		toastType: Info,
		id:        uuid.New(),
	}
	for _, opt := range opts {
		opt(&toast)
	}
	return toast
}

type Toast struct {
	text      string
	toastType toastType
	dur       time.Duration
	id        uuid.UUID
}

func (model *Toast) Text() string {
	return model.text
}

func (model *Toast) ID() uuid.UUID {
	return model.id
}

func (model *Toast) Equals(other *Toast) bool {
	return model.id == other.id
}

const maxWidth = 20

var generalStyle = lipgloss.NewStyle().Width(maxWidth).Inline(false).Bold(true)

var infoStyle = generalStyle.Background(lipgloss.Color("#2b7"))
var warnStyle = generalStyle.Background(lipgloss.Color("#cc2"))
var errorStyle = generalStyle.Background(lipgloss.Color("#d36"))

func (model Toast) Init() tea.Cmd {
	return tea.Tick(model.dur, func(t time.Time) tea.Msg {
		return RemoveToastMsg{ID: model.id}
	})
}

func (model Toast) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return model, nil
}

func (model Toast) View() string {
	switch model.toastType {
	case Info:
		return infoStyle.Render(model.text) + "\n"
	case Warn:
		return warnStyle.Render(model.text) + "\n"
	case Error:
		return errorStyle.Render(model.text) + "\n"
	default:
		return ""
	}
}
