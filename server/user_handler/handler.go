package user_handler

import (
	"encoding/json"
	"io"

	"github.com/ross96D/updater/share"
	"github.com/ross96D/updater/share/configuration"
)

type App struct {
	Index int `json:"index"`
	configuration.Application
}

func HandleUserAppsList(w io.Writer) error {
	list := share.Config().Apps

	apps := make([]App, 0, len(list))
	for i, v := range list {
		app := App{
			Index:       i,
			Application: v,
		}
		apps = append(apps, app)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	return enc.Encode(apps)
}
