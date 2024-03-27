package github

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/google/go-github/v60/github"
	"github.com/ross96D/updater/share"
	"github.com/ross96D/updater/share/configuration"
	taskservice "github.com/ross96D/updater/task_service"
)

func HandleGithubWebhook(payload []byte, eventType string) error {
	event, err := github.ParseWebHook(eventType, payload)
	if err != nil {
		return err
	}
	switch event := event.(type) {
	case *github.ReleaseEvent:
		handleReleaseEvent(event)
	default:
		panic(errors.New("unhandled event"))
	}
	return nil
}

func handleReleaseEvent(event *github.ReleaseEvent) error {
	switch *event.Action {
	case "published", "edited":
		return onPublishEdit(event)
	default:
		panic("unhandled action for release event")
	}
}

func onPublishEdit(event *github.ReleaseEvent) error {
	index := slices.IndexFunc(share.Config().Apps, func(e *configuration.Application) bool {
		return *event.Repo.URL == fmt.Sprintf("github.com/%s/%s", e.Owner, e.Repo)
	})
	if index == -1 {
		return errors.New("release event from repo not configured")
	}
	app := share.Config().Apps[index]
	release := event.Release
	for _, asset := range release.Assets {
		if app.AssetName == *asset.Name {
			handleAssetMatch(app, asset)
		}
	}

	return nil
}

func handleAssetMatch(app *configuration.Application, asset *github.ReleaseAsset) {
	client := github.NewClient(nil).WithAuthToken(app.GithubAuthToken)
	rc, lenght, err := downloadableAsset(client, *asset.URL)
	if err != nil {
		// TODO how to handle
		panic(err)
	}

	path := filepath.Join(*share.Config().BasePath, *asset.Name)
	if err = share.CreateFile(rc, lenght, path); err != nil {
		// TODO how to handle
		panic(err)
	}

	if err = taskservice.Stop(app.TaskSchedPath); err != nil {
		// TODO how to handle
		panic(err)
	}

	if err = os.Rename(app.AppPath, path); err != nil {
		// TODO how to handle
		panic(err)
	}

	if err = taskservice.Start(app.TaskSchedPath); err != nil {
		// TODO how to handle
		panic(err)
	}
	rc.Close()
}

// func updateTaskSched(_ context.Context, rc io.ReadCloser, taskSchedPath string, appPath string) error {
// 	return nil
// }
