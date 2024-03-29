package user_handler

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/ross96D/updater/share"
)

func HandlerUserUpdate(payload []byte) error {
	var app App
	err := json.Unmarshal(payload, &app)
	if err != nil {
		return err
	}
	list := share.Config().Apps
	if app.Index >= len(list) {
		return errors.New("invalid index")
	}
	application := list[app.Index]

	share.HandleAssetMatch(application)

	return nil
}

type App struct {
	Index     int    `json:"index"`
	Host      string `json:"host"`
	Owner     string `json:"owner"`
	Repo      string `json:"repo"`
	AssetName string `json:"asset_name"`
}

func HandleUserAppsList(w io.Writer) error {
	list := share.Config().Apps

	apps := make([]App, 0, len(list))
	for i, v := range list {
		app := App{
			Index:     i,
			Host:      v.Host,
			Owner:     v.Owner,
			Repo:      v.Repo,
			AssetName: v.AssetName,
		}
		apps = append(apps, app)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	return enc.Encode(apps)
}
