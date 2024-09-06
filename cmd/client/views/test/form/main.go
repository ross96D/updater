package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/components/form"
)

func main() {
	f := form.NewForm(
		[][]form.Item{
			{form.Label("key1large"), form.Input[int]()},
			{form.Label("key2"), form.Input[uint]()},
			{form.Label("key3"), form.Input[string]()},
			{form.Label("key4"), form.Input[float64]()},
		},
	)

	_, err := tea.NewProgram(f).Run()
	if err != nil {
		panic(err)
	}
}
