package share

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/go-github/v60/github"
	"github.com/ross96D/updater/share/configuration"
	taskservice "github.com/ross96D/updater/task_service"
)

func verifyChecksum(checksum []byte, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	verify, err := VerifyWithChecksum(checksum, f, sha256.New()) // TODO make the hash algorithm be configurable
	if err != nil {
		return err
	}
	if !verify {
		return errors.New("asset could not be verified")
	}
	return nil
}

func HandleAssetMatch(app *configuration.Application, asset *github.ReleaseAsset, release *github.RepositoryRelease) error {
	client := NewGithubClient(app, nil)
	rc, lenght, err := downloadableAsset(client, *asset.URL)
	if err != nil {
		return err
	}
	defer rc.Close()

	path := filepath.Join(*Config().BasePath, *asset.Name)
	if err = CreateFile(rc, lenght, path); err != nil {
		return err
	}

	var verify = true
	checksum, err := GetChecksum(app, release)
	if err != nil {
		if err != ErrNoChecksum {
			verify = false
		} else {
			return err
		}
	}

	if verify {
		err = verifyChecksum(checksum, path)
		if err != nil {
			return err
		}
	}

	if err = taskservice.Stop(app.TaskSchedPath); err != nil {
		return err
	}

	if err = os.Rename(app.AppPath, app.AppPath+".old"); err != nil {
		return err
	}

	if err = os.Rename(path, app.AppPath); err != nil {
		return err
	}

	if err = taskservice.Start(app.TaskSchedPath); err != nil {
		return err
	}

	return nil
}

func downloadableAsset(ghclient *github.Client, url string) (rc io.ReadCloser, lenght int64, err error) {
	client := ghclient.Client()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Accept", "application/octet-stream")

	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode >= 400 {
		err = errors.New("invalid status code")
		return
	}
	if resp.ContentLength < 0 {
		if resp.ContentLength, err = getHeaders(ghclient, url); err != nil {
			err = fmt.Errorf("head request: %w", err)
			return
		}
	}
	return resp.Body, resp.ContentLength, nil
}

func getHeaders(ghclient *github.Client, url string) (lenght int64, err error) {
	client := ghclient.Client()
	req, err := http.NewRequest(http.MethodHead, url, nil)
	req.Header.Set("Accept", "application/octet-stream")
	if err != nil {
		return
	}
	resp, err := client.Do(req)
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
