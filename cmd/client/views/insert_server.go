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

type EditServerMsg struct {
	server models.Server
	index  int
}

type EditServer EditServerMsg

var EditServerCmd = func(index int, server models.Server) tea.Cmd {
	return func() tea.Msg {
		return EditServerMsg{
			server: server,
			index:  index,
		}
	}
}

const (
	servername = "ServerName"
	address    = "Address"
	username   = "UserName"
	password   = "Password"
)

func NewServerFormView(es *EditServer) ServerFormView {
	sfv := ServerFormView{
		index: -1,
		form: form.NewForm(
			[][]form.Item{
				form.Link(form.Label(servername), form.Input[string]()),
				form.Link(form.Label(address), form.Input(form.WithValidationFromType[*url.URL, models.URLValidator]())),
				form.Link(form.Label(username), form.Input[string]()),
				form.Link(form.Label(password), form.Input(form.WithValidationFromType[models.Password, models.PasswordValidator]())),
			},
		),
	}
	if es != nil {
		sfv.index = es.index
		sfv.Server = es.server
		sfv.setValues()
	}
	return sfv
}

type ServerFormView struct {
	Server models.Server
	form   form.Form
	index  int
}

func (sfv ServerFormView) Init() tea.Cmd {
	return sfv.form.Init()
}

func (sfv ServerFormView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case form.SubmitMsg:
		sfv.fill()
		if !sfv.Validate() {
			return sfv, nil
		}
		if sfv.index == -1 {
			return sfv, tea.Sequence(components.NavigatorPop, InsertServerCmd(sfv.Server))
		} else {
			return sfv, tea.Sequence(components.NavigatorPop, EditServerCmd(sfv.index, sfv.Server))
		}
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
	return !reflect.ValueOf(sfv.Server).IsZero()
}

func (sfv *ServerFormView) fill() {
	if item, ok := sfv.form.GetLinkedValue(servername); ok {
		sfv.Server.ServerName = item.(*form.ItemInput[string]).Value()
	}
	if item, ok := sfv.form.GetLinkedValue(address); ok {
		sfv.Server.Url = item.(*form.ItemInput[*url.URL]).Value()
	}
	if item, ok := sfv.form.GetLinkedValue(password); ok {
		sfv.Server.Password = item.(*form.ItemInput[models.Password]).Value()
	}
	if item, ok := sfv.form.GetLinkedValue(username); ok {
		sfv.Server.UserName = item.(*form.ItemInput[string]).Value()
	}
}

func (sfv *ServerFormView) setValues() {
	if item, ok := sfv.form.GetLinkedValue(servername); ok {
		item.(*form.ItemInput[string]).SetValue(sfv.Server.ServerName)
		item.(*form.ItemInput[string]).SetText(sfv.Server.ServerName)
	}
	if item, ok := sfv.form.GetLinkedValue(address); ok {
		item.(*form.ItemInput[*url.URL]).SetValue(sfv.Server.Url)
		item.(*form.ItemInput[*url.URL]).SetText(sfv.Server.Url.String())
	}
	if item, ok := sfv.form.GetLinkedValue(password); ok {
		item.(*form.ItemInput[models.Password]).SetValue(sfv.Server.Password)
		item.(*form.ItemInput[models.Password]).SetText(string(sfv.Server.Password))
	}
	if item, ok := sfv.form.GetLinkedValue(username); ok {
		item.(*form.ItemInput[string]).SetValue(sfv.Server.UserName)
		item.(*form.ItemInput[string]).SetText(sfv.Server.UserName)
	}
}
