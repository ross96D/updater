package form

import (
	"reflect"
	"strconv"
	"unsafe"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/ross96D/updater/cmd/client/components/label"
)

func Link[T any](label ItemLabel, input ItemInput[T]) []Item {
	input.linkID = label.ID()
	return []Item{label, input}
}

var styleFocus = lipgloss.NewStyle().Background(lipgloss.Color("#0bb"))
var styleEmpty = lipgloss.NewStyle()
var sytleErr = lipgloss.NewStyle().Background(lipgloss.Color("#d22"))

func generateID() uint32 {
	return uuid.New().ID()
}

type itemType int

const (
	inputType itemType = iota + 1
	labelType
)

type Item interface {
	ID() uint32
	getType() itemType
	View() string
	Update(tea.Msg) (Item, tea.Cmd)
	Blur() Item
	Focus() Item
	linkedId() uint32
}

// Form item that represent a label
type ItemLabel struct {
	name    string
	label   label.Label
	id      uint32
	isFocus bool
}

func (item ItemLabel) Blur() Item {
	item.isFocus = false
	return item
}

func (item ItemLabel) Focus() Item {
	item.isFocus = true
	return item
}

func (f ItemLabel) linkedId() uint32 {
	return 0
}

func (f ItemLabel) ID() uint32 {
	return f.id
}

func (f ItemLabel) Update(msg tea.Msg) (Item, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		f.label.Update(msg)
	}
	if _, ok := msg.(acceptInputMsg); ok {
		return f, NextItemCmd
	}
	return f, nil
}

func (ItemLabel) getType() itemType {
	return labelType
}

func (f ItemLabel) View() string {
	style := &styleEmpty
	if f.isFocus {
		style = &styleFocus
	}

	return style.Render(f.label.Render())
}

func Label(text string) ItemLabel {
	return ItemLabel{
		name: text,
		id:   generateID(),
		label: label.NewText(
			text,
			label.TextStyle(
				lipgloss.NewStyle().
					PaddingLeft(1).
					PaddingRight(2),
			),
		),
	}
}

// for parsing data in input values
type ErrValidation string

func NewErrValidation(err error) error {
	if err != nil {
		return ErrValidation(err.Error())
	}
	return nil
}

type errValidationMsg struct {
	ErrValidation
	id uint32
}

func errValidationCmd(item Item, err ErrValidation) tea.Cmd {
	return func() tea.Msg {
		return errValidationMsg{
			id:            item.ID(),
			ErrValidation: err,
		}
	}
}

func (err ErrValidation) Error() string {
	return string(err)
}

type ParseValidation[T any] func(string) (T, error)

type CustomParseValidation[T any] interface {
	ParseValidationItem(string) (T, error)
}

type acceptInputMsg struct{}

// Form item that for an input field
type ItemInput[T any] struct {
	value        T
	parse        ParseValidation[T]
	onAccept     func() tea.Cmd
	errorMessage string
	label        label.Label
	input        textinput.Model
	id           uint32
	linkID       uint32
	isFocus      bool
}

func (item ItemInput[T]) Blur() Item {
	item.input.Blur()
	item.isFocus = false
	return item
}

func (item ItemInput[T]) Focus() Item {
	item.input.Focus()
	item.isFocus = true
	return item
}

func (item ItemInput[T]) linkedId() uint32 {
	return item.linkID
}

func (ItemInput[T]) getType() itemType {
	return inputType
}

func (f ItemInput[T]) View() string {
	wrapper := &styleEmpty
	if f.errorMessage != "" {
		wrapper = &sytleErr
	} else if f.isFocus {
		wrapper = &styleFocus
	}
	return wrapper.Render(lipgloss.JoinHorizontal(lipgloss.Left, f.label.Render(), f.input.View()))
}

func (f ItemInput[T]) ID() uint32 {
	return f.id
}

func (f ItemInput[T]) Update(msg tea.Msg) (Item, tea.Cmd) {
	switch msg := msg.(type) {
	case acceptInputMsg:
		var err error
		f.value, err = f.parse(f.input.Value())
		if err != nil {
			return f, errValidationCmd(f, err.(ErrValidation))
		}

		var cmd tea.Cmd
		cmd = NextItemCmd
		if f.onAccept != nil {
			cmd = tea.Batch(cmd, f.onAccept())
		}
		return f, cmd

	case tea.WindowSizeMsg:
		f.label.Update(msg)
		return f, nil

	case tea.KeyMsg:
		f.errorMessage = ""

	case errValidationMsg:
		f.errorMessage = string(msg.ErrValidation)
	}

	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)

	return f, cmd
}

func (item ItemInput[T]) Value() T {
	return item.value
}

type inputOptions[T any] func(*ItemInput[T])

func WithValidation[T any](parser ParseValidation[T]) inputOptions[T] {
	return func(ii *ItemInput[T]) {
		ii.parse = parser
	}
}

func WithValidationFromType[T any, V CustomParseValidation[T]]() inputOptions[T] {
	return func(ii *ItemInput[T]) {
		ii.parse = (*new(V)).ParseValidationItem
	}
}

func WithOnAccept[T any](onAccept func() tea.Cmd) inputOptions[T] {
	return func(ii *ItemInput[T]) {
		ii.onAccept = onAccept
	}
}

func Input[T any](opts ...inputOptions[T]) ItemInput[T] {
	item := ItemInput[T]{
		id:    generateID(),
		input: textinput.New(),
	}
	for _, opt := range opts {
		opt(&item)
	}

	validationFallback(&item)

	if item.parse == nil {
		panic("no parse function placed")
	}
	return item
}

func unsafeCast[T, V any](v V) T {
	return *(*T)(unsafe.Pointer(&v))
}

func validationFallback[T any](i *ItemInput[T]) {
	if i.parse != nil {
		return
	}

	temp := any(*new(T))

	if t, ok := temp.(CustomParseValidation[T]); ok {
		i.parse = t.ParseValidationItem
		return
	}

	switch temp.(type) {
	case int:
		i.parse = func(s string) (T, error) {
			v, err := strconv.Atoi(s)
			err = NewErrValidation(err)
			sd := unsafeCast[T](v)
			return sd, err
		}
	case int8:
		i.parse = func(s string) (T, error) {
			v, err := strconv.ParseInt(s, 10, 8)
			err = NewErrValidation(err)
			sd := unsafeCast[T](v)
			return sd, err
		}
	case int16:
		i.parse = func(s string) (T, error) {
			v, err := strconv.ParseInt(s, 10, 16)
			err = NewErrValidation(err)
			sd := unsafeCast[T](v)
			return sd, err
		}
	case int32:
		i.parse = func(s string) (T, error) {
			v, err := strconv.ParseInt(s, 10, 32)
			err = NewErrValidation(err)
			sd := unsafeCast[T](v)
			return sd, err
		}
	case int64:
		i.parse = func(s string) (T, error) {
			v, err := strconv.ParseInt(s, 10, 64)
			err = NewErrValidation(err)
			sd := unsafeCast[T](v)
			return sd, err
		}
	case uint:
		i.parse = func(s string) (T, error) {
			v, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				err = NewErrValidation(err)
			}
			return unsafeCast[T](v), err
		}
	case uint8:
		i.parse = func(s string) (T, error) {
			v, err := strconv.ParseUint(s, 10, 8)
			err = NewErrValidation(err)
			return unsafeCast[T](v), err
		}
	case uint16:
		i.parse = func(s string) (T, error) {
			v, err := strconv.ParseUint(s, 10, 16)
			err = NewErrValidation(err)
			return unsafeCast[T](v), err
		}
	case uint32:
		i.parse = func(s string) (T, error) {
			v, err := strconv.ParseUint(s, 10, 32)
			err = NewErrValidation(err)
			return unsafeCast[T](v), err
		}
	case uint64:
		i.parse = func(s string) (T, error) {
			v, err := strconv.ParseUint(s, 10, 64)
			err = NewErrValidation(err)
			return unsafeCast[T](v), err
		}
	case float32:
		i.parse = func(s string) (T, error) {
			v, err := strconv.ParseFloat(s, 32)
			err = NewErrValidation(err)
			return unsafeCast[T](v), err
		}
	case float64:
		i.parse = func(s string) (T, error) {
			v, err := strconv.ParseFloat(s, 64)
			err = NewErrValidation(err)
			return unsafeCast[T](v), err
		}
	case string:
		i.parse = func(s string) (T, error) {
			return unsafeCast[T](s), nil
		}
	default:
		panic("fallback not found for type " + reflect.TypeOf(temp).String())
	}
}