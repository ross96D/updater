package github_handler

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/google/go-github/v60/github"
	"github.com/ross96D/updater/share"
	"github.com/ross96D/updater/share/configuration"
	"github.com/rs/zerolog/log"
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
		log.Warn().Interface("unhandled event", event).Send()
		return errors.New("unhandled event")
	}
}

func handleReleaseEvent(event *github.ReleaseEvent) error {
	switch *event.Action {
	// case "released", "published", "edited", "created":
	case "published":
		log.Info().Msg("github event action " + *event.Action)
		return onPublishEdit(event)
	default:
		log.Warn().Msg("unhandled action for release event " + *event.Action)
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
		log.Warn().Msg(fmt.Sprintf("release event from repo github.com/%s/%s not configured", eventOwner, eventRepo))
		return errors.New("release event from repo not configured")
	}
	app := share.Config().Apps[index]
	release, err := getRelease(app, *event.Release.ID)
	if err != nil {
		return err
	}
	if len(release.Assets) == 0 {
		log.Info().Msg("release with out assets")
		return errors.New("release with out assets")
	}

	for _, asset := range release.Assets {
		if app.AssetName == *asset.Name {
			return share.HandleAssetMatch(app, asset, release)
		}
	}
	log.Warn().Msg(fmt.Sprintf("asset not found asset to match: %s release assets: %s\n", app.AssetName, share.SingleLineSlice(release.Assets)))
	return errors.New("not found asset match")
}

func getRelease(app configuration.Application, id int64) (*github.RepositoryRelease, error) {
	client := share.NewGithubClient(app, nil)
	resp, _, err := client.Repositories.GetRelease(context.Background(), app.Owner, app.Repo, id)
	// TODO so, if there is an error and the github response exist.. the github response should be added to the error or logged.
	return resp, err
}
