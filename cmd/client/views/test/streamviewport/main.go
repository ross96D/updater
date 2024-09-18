package main

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/andrewstuart/limio"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/components/streamviewport"
	"github.com/ross96D/updater/cmd/client/pretty"
)

func main() {
	pretty.ActivateDebug() //nolint: errcheck
	_, filename, _, _ := runtime.Caller(0)
	err := os.Chdir(filepath.Dir(filename))
	if err != nil {
		panic(err)
	}

	f, err := os.Open("pager_data.md")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	limitReader := limio.NewReader(f)
	limitReader.SimpleLimit(10*limio.KB, time.Minute)
	m := streamviewport.New(limitReader, 80, 20)

	if _, err := tea.NewProgram(m, tea.WithMouseCellMotion()).Run(); err != nil {
		panic(err)
	}
}
