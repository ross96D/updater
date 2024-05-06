package user_handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/google/go-github/v60/github"
	"github.com/ross96D/updater/share"
	"github.com/ross96D/updater/share/configuration"
	"github.com/rs/zerolog/log"
)

func GetReleaseRepository(app configuration.Application) (*github.RepositoryRelease, *github.Response, error) {
	client := share.NewGithubClient(app, nil)
	return client.Repositories.GetLatestRelease(context.TODO(), app.Owner, app.Repo)
}

func HandlerUserUpdate(payload []byte) error {
	var app App
	err := json.Unmarshal(payload, &app)
	if err != nil {
		log.Error().Err(fmt.Errorf("unmarshaling app from user %w", err)).Send()
		return err
	}
	log.Info().Interface("user app", app).Send()
	list := share.Config().Apps
	if app.Index >= len(list) {
		return errors.New("invalid index")
	}
	application := list[app.Index]

	release, _, err := GetReleaseRepository(application)
	if err != nil {
		log.Error().Err(fmt.Errorf("getReleaseRepository from user %w", err)).Send()
		return err
	}

	return share.UpdateApp(application, release)
}

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
