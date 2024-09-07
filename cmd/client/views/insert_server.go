package views

import (
	"net/url"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ross96D/updater/cmd/client/components"
	"github.com/ross96D/updater/cmd/client/components/form"
	"github.com/ross96D/updater/cmd/client/models"
)

type InsertServerMsg models.Server

var InsertServerCmd = func(s models.Server) tea.Cmd {
	return func() tea.Msg {
		return InsertServerMsg(s)
	}
}

const (
	servername = "ServerName"
	address    = "Address"
	username   = "UserName"
	password   = "Password"
)

func NewServerFormView() ServerFormView {
	return ServerFormView{
		form: form.NewForm(
			[][]form.Item{
				form.Link(form.Label(servername), form.Input[string]()),
				form.Link(form.Label(address), form.Input(form.WithValidationFromType[*url.URL, models.URLValidator]())),
				form.Link(form.Label(username), form.Input[string]()),
				form.Link(form.Label(password), form.Input(form.WithValidationFromType[models.Password, models.PasswordValidator]())),
			},
		),
	}
}

type ServerFormView struct {
	Server models.Server
	form   form.Form
}

func (sfv ServerFormView) Init() tea.Cmd {
	return sfv.form.Init()
}

func (sfv ServerFormView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case form.SubmitMsg:
		sfv.fill()
		// TODO validate
		if reflect.ValueOf(sfv.Server).IsZero() {
			panic("ZERO")
		}
		return sfv, tea.Sequence(components.NavigatorPop, InsertServerCmd(sfv.Server))
	}
	model, cmd := sfv.form.Update(msg)
	sfv.form = model.(form.Form)
	return sfv, cmd
}

func (sfv ServerFormView) View() string {
	return sfv.form.View()
}

// TODO implement validation
func (sfv *ServerFormView) Validate() bool {
	return true
}

func (sfv *ServerFormView) fill() {
	if item, ok := sfv.form.GetLinkedValue(servername); ok {
		sfv.Server.ServerName = item.(form.ItemInput[string]).Value()
	}
	if item, ok := sfv.form.GetLinkedValue(address); ok {
		sfv.Server.Url = item.(form.ItemInput[*url.URL]).Value()
	}
	if item, ok := sfv.form.GetLinkedValue(password); ok {
		sfv.Server.Password = item.(form.ItemInput[models.Password]).Value()
	}
	if item, ok := sfv.form.GetLinkedValue(username); ok {
		sfv.Server.UserName = item.(form.ItemInput[string]).Value()
	}
}
