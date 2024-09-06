package label

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LabelOptions func(*Label)

// Prefer the use of this struct for static text views
type Label struct {
	text  string
	style lipgloss.Style

	wrapped bool
	width   int
}

func TextStyle(style lipgloss.Style) LabelOptions {
	return func(t *Label) {
		t.style = style
	}
}

func TextWrapped(wrapped bool) LabelOptions {
	return func(t *Label) {
		t.wrapped = wrapped
	}
}

func TextWidth(w int) LabelOptions {
	return func(t *Label) {
		t.width = w
	}
}

func NewText(text string, opts ...LabelOptions) Label {
	t := Label{text: text}
	for _, v := range opts {
		v(&t)
	}
	return t
}

func (t *Label) Update(msg tea.WindowSizeMsg) {
	if !t.wrapped {
		return
	}

	t.width = msg.Width
}

func (t Label) Render() string {
	if t.wrapped {
		return t.style.Width(t.width).Render(t.text)
	}

	return t.style.Render(t.text)
}
