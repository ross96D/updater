package toast

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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
	return Toast{}
}

type Toast struct {
	text      string
	toastType toastType
	dur       time.Duration
}

var generalStyle = lipgloss.NewStyle().Height(2).Width(20)

var infoStyle = generalStyle.Background(lipgloss.Color("#2b7"))
var warnStyle = generalStyle.Background(lipgloss.Color("#cc2"))
var errorStyle = generalStyle.Background(lipgloss.Color("#d36"))

func (model Toast) Init() tea.Cmd {
	// tea.Tick()
	return nil
}

func (model Toast) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return model, nil
}

func (model Toast) View() string {
	switch model.toastType {
	case Info:
		return infoStyle.Render(model.text)
	case Warn:
		return warnStyle.Render(model.text)
	case Error:
		return errorStyle.Render(model.text)
	default:
		panic(fmt.Sprintf("unexpected toast.toastType: %#v", model.toastType))
	}
}
