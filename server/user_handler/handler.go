package user_handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/ross96D/updater/logger"
	"github.com/ross96D/updater/share"
	"github.com/ross96D/updater/share/configuration"
	"github.com/ross96D/updater/share/match"
	"github.com/rs/zerolog/log"
)

type GithubReleaseData struct {
	client  *github.Client
	release *github.RepositoryRelease
}

func NewGithubReleaseData(app configuration.Application) (match.Data, error) {
	if app.GithubRelease == nil {
		return nil, errors.New("no github repo configured")
	}

	client := github.NewClient(nil)
	if app.GithubRelease.Token != "" {
		client = client.WithAuthToken(app.GithubRelease.Token)
	}

	release, _, err := client.Repositories.GetLatestRelease(context.TODO(), app.GithubRelease.Owner, app.GithubRelease.Repo)
	if err != nil {
		return nil, fmt.Errorf("NewGithubReleaseData GetLatestRelease() %w", err)
	}

	return GithubReleaseData{client: client, release: release}, nil
}

func (gd GithubReleaseData) Clean() {}
func (gd GithubReleaseData) Get(name string) io.ReadCloser {
	if name == "" {
		return nil
	}
	var toDownload *github.ReleaseAsset
	for _, v := range gd.release.Assets {
		if v.Name != nil && *v.Name == name {
			toDownload = v
			break
		}
	}
	if toDownload == nil {
		return nil
	}

	rc, _, err := downloadableAsset(gd.client, *toDownload.URL)
	if err != nil {
		log.Error().Err(err).Msg("error in GithubReleaseData downloadableAsset()")
		return nil
	}
	return rc
}

func downloadableAsset(client *github.Client, url string) (rc io.ReadCloser, lenght int64, err error) {
	req, err := client.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Accept", "application/octet-stream")

	if err != nil {
		return
	}
	resp, err := client.BareDo(context.TODO(), req)
	if err != nil {
		return
	}
	if resp.StatusCode >= 400 {
		err = errors.New("invalid status code")
		return
	}
	if resp.ContentLength < 0 {
		if resp.ContentLength, err = getHeaders(client, url); err != nil {
			err = fmt.Errorf("head request: %w", err)
			return
		}
	}
	return resp.Body, resp.ContentLength, nil
}

func getHeaders(client *github.Client, url string) (lenght int64, err error) {
	req, err := client.NewRequest(http.MethodHead, url, nil)
	req.Header.Set("Accept", "application/octet-stream")
	if err != nil {
		return
	}
	resp, err := client.BareDo(context.TODO(), req)
	if err != nil {
		return
	}
	if resp.StatusCode >= 400 {
		err = errors.New("invalid status code")
		return
	}
	lenght = resp.ContentLength
	return
}

func GetReleaseRepository(app configuration.Application) (*github.RepositoryRelease, *github.Response, error) {
	client := github.NewClient(nil)
	if app.GithubRelease.Token != "" {
		client = client.WithAuthToken(app.GithubRelease.Token)
	}
	return client.Repositories.GetLatestRelease(context.TODO(), app.GithubRelease.Owner, app.GithubRelease.Repo)
}

func HandlerUserUpdate(ctx context.Context, payload []byte, dryRun bool) error {
	var app App
	err := json.Unmarshal(payload, &app)
	if err != nil {
		return fmt.Errorf("HandlerUserUpdate Unmarshall() %w", err)
	}
	log.Info().Interface("user app", app).Send()
	list := share.Config().Apps
	if app.Index >= len(list) {
		return match.NewErrError(errors.New("HandlerUserUpdate invalid index"))
	}
	application := list[app.Index]
	if application.GithubRelease == nil {
		return errors.New("no github repo configured")
	}

	logger, _ := logger.LoggerCtx_FromContext(ctx)
	logger.Info().Msgf("Requesting release from github.com/%s/%s ", application.GithubRelease.Owner, application.GithubRelease.Repo)
	var data match.Data
	if !dryRun {
		data, err = NewGithubReleaseData(application)
		if err != nil {
			logger.Info().Msg("Requesting release failed")
			return match.NewErrError(err)
		}
	} else {
		data = match.EmptyData{}
	}
	time.Sleep(time.Second)
	return match.Update(ctx, application, match.WithData(data), match.WithDryRun(dryRun))
}

type Server struct {
	Apps    []App             `json:"apps"`
	Version share.VersionData `json:"version"`
}

type App struct {
	configuration.Application
	Index int `json:"index"`
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

	return enc.Encode(Server{Apps: apps, Version: share.Version()})
}
