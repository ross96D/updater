package github_handler

import (
	"errors"
	"fmt"
	"log"
	"slices"

	"github.com/google/go-github/v60/github"
	"github.com/ross96D/updater/share"
	"github.com/ross96D/updater/share/configuration"
)

func HandleGithubWebhook(payload []byte, eventType string) error {
	event, err := github.ParseWebHook(eventType, payload)
	if err != nil {
		return err
	}
	switch event := event.(type) {
	case *github.ReleaseEvent:
		return handleReleaseEvent(event)
	default:
		log.Printf("unhandled event %+v\n", event)
		return errors.New("unhandled event")
	}
}

func handleReleaseEvent(event *github.ReleaseEvent) error {
	switch *event.Action {
	case "published", "edited":
		return onPublishEdit(event)
	default:
		log.Printf("unhandled action for release event %s\n", *event.Action)
		return errors.New("unhandled action for release event")
	}
}

func onPublishEdit(event *github.ReleaseEvent) error {
	index := slices.IndexFunc(share.Config().Apps, func(e configuration.Application) bool {
		return *event.Repo.URL == fmt.Sprintf("github.com/%s/%s", e.Owner, e.Repo)
	})
	if index == -1 {
		return errors.New("release event from repo not configured")
	}
	app := share.Config().Apps[index]
	release := event.Release
	for _, asset := range release.Assets {
		if app.AssetName == *asset.Name {
			return share.HandleAssetMatch(app, asset, release)
		}
	}

	return nil
}
