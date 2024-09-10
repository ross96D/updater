package form

import (
	"iter"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SubmitMsg struct{}

var SubmitCmd = func() tea.Msg { return SubmitMsg{} }

type NextItemMsg struct{}

var NextItemCmd = func() tea.Msg { return NextItemMsg{} }

func NewForm(items [][]Item) Form {
	// TODO there is a bug if first row does not have any item
	if len(items) > 0 && len(items[0]) > 0 {
		items[0][0] = items[0][0].Focus()
	}
	return Form{
		items: items,
	}
}

type Form struct {
	items    [][]Item
	focusRow int
	focusCol int
}

func (f Form) Init() tea.Cmd {
	return nil
}

func (f Form) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msgT := msg.(type) {
	case tea.KeyMsg:
		switch msgT.Type {
		case tea.KeyCtrlC:
			return f, tea.Quit

		// navigation
		case tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight:
			f.updateFocus(msgT)
			return f, nil

		// accept value of input
		case tea.KeyEnter:
			var cmd tea.Cmd
			f.items[f.focusRow][f.focusCol], cmd = f.items[f.focusRow][f.focusCol].Update(acceptInputMsg{})
			return f, cmd

		// Remap keyShift<Direction> to key<Direction>
		case tea.KeyShiftUp:
			msg = tea.Key{Type: tea.KeyUp}
		case tea.KeyShiftDown:
			msg = tea.Key{Type: tea.KeyDown}
		case tea.KeyShiftLeft:
			msg = tea.Key{Type: tea.KeyLeft}
		case tea.KeyShiftRight:
			msg = tea.Key{Type: tea.KeyRight}
		}

	case errValidationMsg:
		f.addError(msgT)
		return f, nil

	case NextItemMsg:
		f.nextItem()
		return f, nil
	}
	var cmd tea.Cmd
	f.items[f.focusRow][f.focusCol], cmd = f.items[f.focusRow][f.focusCol].Update(msg)

	return f, cmd
}

func (f *Form) nextItem() {
	prev := struct {
		row int
		col int
	}{
		row: f.focusRow,
		col: f.focusCol,
	}
	if ok := f.right(); !ok {
		if ok = f.down(); ok {
			f.focusCol = 0
		}
	}
	actual := struct {
		row int
		col int
	}{
		row: f.focusRow,
		col: f.focusCol,
	}
	if prev != actual {
		f.items[prev.row][prev.col] = f.items[prev.row][prev.col].Blur()
		f.items[f.focusRow][f.focusCol] = f.items[f.focusRow][f.focusCol].Focus()
	}
}

func (f *Form) addError(msg errValidationMsg) {
	item := f.find(msg.id)
	if item == nil {
		return
	}
	*item, _ = (*item).Update(msg)
}

func (f *Form) find(id uint32) *Item {
	for _, row := range f.items {
		for j := range row {
			if row[j].ID() == id {
				return &row[j]
			}
		}
	}
	return nil
}

func (f *Form) findLabel(name string) (*ItemLabel, bool) {
	for _, row := range f.items {
		for j := range row {
			if item, ok := row[j].(*ItemLabel); ok && item.name == name {
				return item, true
			}
		}
	}
	return nil, false
}

func (f *Form) findInputByLink(link uint32) Item {
	for _, row := range f.items {
		for _, item := range row {
			if id := item.linkedId(); id != 0 && id == link {
				return item
			}
		}
	}
	return nil
}

func (f *Form) up() bool {
	if f.focusRow == 0 {
		return false
	}
	f.focusRow -= 1
	return true
}

func (f *Form) down() bool {
	lenRows := len(f.items)

	if f.focusRow == lenRows-1 {
		return false
	}
	f.focusRow += 1
	return true
}

func (f *Form) right() bool {
	lenCols := len(f.items[0])

	if f.focusCol == lenCols-1 {
		return false
	}
	f.focusCol += 1
	return true
}

func (f *Form) left() bool {
	if f.focusCol == 0 {
		return false
	}
	f.focusCol -= 1
	return true
}

func (f *Form) updateFocus(key tea.KeyMsg) {
	prev := struct {
		row int
		col int
	}{row: f.focusRow, col: f.focusCol}

	switch key.Type {
	case tea.KeyDown:
		f.down()

	case tea.KeyUp:
		f.up()

	case tea.KeyRight:
		f.right()

	case tea.KeyLeft:
		f.left()

	default:
		return
	}
	f.items[prev.row][prev.col] = f.items[prev.row][prev.col].Blur()
	f.items[f.focusRow][f.focusCol] = f.items[f.focusRow][f.focusCol].Focus()
}

func (f Form) View() string {
	transform := func(_ int, row Item) string {
		return row.View()
	}

	cols := make([]string, 0)

	for i := range f.items[0] {
		ss := collect(_map(f.Cols(i), transform))
		cols = append(cols, lipgloss.PlaceVertical(len(ss), lipgloss.Left, strings.Join(ss, "\n")))
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, cols...)
}

func (f Form) GetLinkedValue(labelName string) (any, bool) {
	label, ok := f.findLabel(labelName)
	if !ok {
		return nil, false
	}
	input := f.findInputByLink(label.id)
	return input, input != nil
}

func (f *Form) Cols(col int) iter.Seq2[int, Item] {
	return func(yield func(int, Item) bool) {
		for rowIndex := 0; rowIndex < len(f.items); rowIndex++ {
			if !yield(rowIndex, f.items[rowIndex][col]) {
				return
			}
		}
	}
}

func (f *Form) Rows(row int) iter.Seq2[int, Item] {
	return func(yield func(int, Item) bool) {
		for colIndex := 0; colIndex < len(f.items[0]); colIndex++ {
			if !yield(colIndex, f.items[row][colIndex]) {
				return
			}
		}
	}
}

func _map[T, V any](seq iter.Seq2[int, T], transform func(int, T) V) iter.Seq2[int, V] {
	gen := func(yield func(int, V) bool) {
		next, _ := iter.Pull2(seq)
		for {
			i, v, ok := next()
			if !ok {
				return
			}
			if !yield(i, transform(i, v)) {
				return
			}
		}
	}
	return gen
}

func collect[T any](seq iter.Seq2[int, T]) []T {
	result := make([]T, 0)
	next, _ := iter.Pull2(seq)

	for {
		_, v, ok := next()
		if !ok {
			break
		}
		result = append(result, v)
	}
	return result
}
