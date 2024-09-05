package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ross96D/updater/cmd/client/models"
	"github.com/ross96D/updater/share/configuration"
)

type ServerView struct {
	Server models.Server
	left   viewport.Model
	rigth  viewport.Model
}

func NewServerView(server models.Server) ServerView {
	left := viewport.New(30, 5)
	rigth := viewport.New(30, 5)

	sv := ServerView{
		Server: server,
		left:   left,
		rigth:  rigth,
	}
	sv.left.SetContent(lipgloss.NewStyle().Width(30).Align(lipgloss.Center).Render(server.Name))
	sv.rigth.SetContent(sv.renderApps())
	return sv
}

func (ServerView) Init() tea.Cmd {
	return nil
}

func (sv ServerView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return sv, tea.Quit
		}
	}

	var cmdLeft tea.Cmd
	var cmdRigth tea.Cmd
	sv.left, cmdLeft = sv.left.Update(msg)
	sv.rigth, cmdRigth = sv.rigth.Update(msg)

	return sv, tea.Batch(cmdLeft, cmdRigth)
}

func (sv ServerView) View() string {
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Render(lipgloss.JoinHorizontal(lipgloss.Center, sv.left.View(), sv.rigth.View()))
}

func (sv ServerView) renderApps() string {
	builder := strings.Builder{}
	builder.WriteString("IP: " + sv.Server.IP + "\n")
	builder.WriteString("apps:\n")

	ident := "\t"

	for _, app := range sv.Server.Apps {
		builder.WriteString(ident)
		builder.WriteString(fmt.Sprintf("Index %d %s", app.Index, app.AuthToken))
		builder.WriteRune('\n')
		builder.WriteString(ident)
		builder.WriteString("assets:")
		builder.WriteRune('\n')
		for _, asset := range app.Assets {
			builder.WriteString(ident + ident)
			builder.WriteString(renderAsset(asset))
			builder.WriteRune('\n')
		}

	}
	return builder.String()
}

func renderAsset(asset configuration.Asset) string {
	return asset.Name
}
