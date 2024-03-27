package server

import (
	"errors"
	"fmt"
	"slices"

	"github.com/google/go-github/v60/github"
	"github.com/ross96D/updater/share"
	"github.com/ross96D/updater/share/configuration"
)

func handleGithubWebhook(payload []byte, eventType string) error {
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
		index := slices.IndexFunc(share.Config().Events, func(e *configuration.AcceptedEvents) bool {
			return *event.Repo.URL == fmt.Sprintf("github.com/%s/%s", e.Owner, e.Repo)
		})
		if index == -1 {
			return errors.New("release event from repo not configured")
		}

		release := event.Release
		for _, asset := range release.Assets {
			contains := slices.ContainsFunc(
				share.Config().Events[index].DownloadableAssests,
				func(e string) bool {
					return e == *asset.Name
				},
			)
			if contains {
				// Download asset
			}
		}
	default:
		panic("unhandled action for release event")
	}

	return nil
}
