package github_handler

import (
	"errors"
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
	case "released", "published", "edited", "created":
		return onPublishEdit(event)
	default:
		log.Printf("unhandled action for release event %s\n", *event.Action)
		return errors.New("unhandled action for release event")
	}
}

func onPublishEdit(event *github.ReleaseEvent) error {
	var eventRepo string = *event.Repo.Name
	var eventOwner string = *event.Repo.Owner.Login
	index := slices.IndexFunc(share.Config().Apps, func(e configuration.Application) bool {
		return eventOwner == e.Owner && eventRepo == e.Repo
	})
	if index == -1 {
		log.Printf("release event from repo github.com/%s/%s not configured", eventOwner, eventRepo)
		return errors.New("release event from repo not configured")
	}
	app := share.Config().Apps[index]
	release := event.Release
	for _, asset := range release.Assets {
		if app.AssetName == *asset.Name {
			return share.HandleAssetMatch(app, asset, release)
		}
	}
	log.Printf("asset not found asset to match: %s release assets: %s\n", app.AssetName, share.SingleLineSlice(release.Assets))
	return nil
}
