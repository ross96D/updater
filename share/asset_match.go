package share

import (
	"context"
	"errors"
	"fmt"
	"hash"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/google/go-github/v60/github"
	"github.com/ross96D/updater/share/configuration"
	taskservice "github.com/ross96D/updater/task_service"
)

var ErrIsChached = fmt.Errorf("asset is cached")
var ErrUnverifiedAsset = fmt.Errorf("unverified asset")
var _mutexHandleAssetMatch = sync.Mutex{}

func HandleAssetMatch(
	app configuration.Application,
	asset *github.ReleaseAsset,
	release *github.RepositoryRelease,
) error {
	// get the checksum
	var verify = true
	checksum, err := GetChecksum(app, app.Checksum, release)
	if err != nil {
		if err == ErrNoChecksum {
			verify = false
		} else {
			return err
		}
	}
	log.Println("should verify the asset", verify)

	// if there is a checksum, verify that the app is not the same as the one on the repo using the hashes
	// if there is no checksum then the cache verification will be performed later on the function
	if verify && app.UseCache && cacheWithChecksum(checksum, app) {
		return ErrIsChached
	}

	log.Println("Donwloading asset")
	rc, lenght, err := downloadableAsset(NewGithubClient(app, nil), *asset.URL)
	if err != nil {
		return err
	}

	tempPath := filepath.Join(Config().BasePath, *asset.Name)
	log.Println("save asset to temporary file in", tempPath)
	if tempPath, err = CreateFile(rc, lenght, tempPath); err != nil {
		return err
	}
	// remove temp file
	defer func() {
		log.Println("removing temporary file")
		os.Remove(tempPath)
	}()

	if verify {
		log.Println("verifiying checksum")
		// verifiy that the checksum correspond to the downloaded asset
		if !verifyChecksum(checksum, tempPath) {
			return ErrUnverifiedAsset
		}
	} else {
		log.Println("checking if asset is already installed")
		// as there is no checksum hash the both, the app file and the downloaded one.
		if cacheWithFile(tempPath, app) {
			return ErrIsChached
		}
	}

	additionalAssetsTempPath, err := CreateAdditionalAssets(app, release)
	if err != nil {
		log.Println("error downloading assets", err.Error())
	}

	log.Println("stoping task")
	if err = taskservice.Stop(app.TaskSchedPath); err != nil {
		return err
	}
	defer func() {
		log.Println("Run the task")
		if err := taskservice.Start(app.TaskSchedPath); err != nil {
			log.Println("Error reruning the task", err.Error())
		}
	}()

	_mutexHandleAssetMatch.Lock()
	defer _mutexHandleAssetMatch.Unlock()
	log.Println("Moving app to app.old")
	if err = RenameSafe(app.SystemPath, app.SystemPath+".old"); err != nil {
		return err
	}

	log.Println("Moving asset to app path")
	if err = Copy(tempPath, app.SystemPath); err != nil {
		// Roll back
		log.Println("Error: copy %s to %s. err: %s", TempPath, app.SystemPath, err.Error())
		os.Remove(app.SystemPath)
		RenameSafe(app.SystemPath+".old", app.SystemPath)
		return err
	}

	for _, p := range additionalAssetsTempPath {
		if err = RenameSafe(p.SystemPath, p.SystemPath+".old"); err != nil {
			// log error
			log.Println("Error: RenameSafe %s to %s. err: %s", p.SystemPath, p.SystemPath+".old", err.Error())
			continue
		}
		if err = Copy(p.TempPath, p.SystemPath); err != nil {
			// Roll back
			log.Println("Error: copy %s to %s. err: %s", p.TempPath, p.SystemPath, err.Error())
			os.Remove(p.SystemPath)
			RenameSafe(p.SystemPath+".old", p.SystemPath)
			return err
		}
	}

	return nil
}

type additionalAssetPath struct {
	SystemPath string
	TempPath   string
}

func CreateAdditionalAssets(
	app configuration.Application,
	release *github.RepositoryRelease,
) ([]additionalAssetPath, error) {
	result := make([]additionalAssetPath, 0, len(app.AdditionalAssets))
	for _, a := range app.AdditionalAssets {
		index := slices.IndexFunc(
			release.Assets,
			func(e *github.ReleaseAsset) bool {
				return *e.Name == a.Name
			},
		)
		if index < 0 {
			continue
		}

		path, err := HandleAdditionalAsset(app, a, release.Assets[index], release)
		if err != nil {
			// error log
			log.Println("error handling additional asset", a.Name, "error", err.Error())
			continue
		}
		result = append(result, additionalAssetPath{
			SystemPath: a.SystemPath,
			TempPath:   path,
		})
	}
	return result, nil
}

func HandleAdditionalAsset(
	app configuration.Application,
	appAsset configuration.AdditionalAsset,
	asset *github.ReleaseAsset,
	release *github.RepositoryRelease,
) (string, error) {
	var verify = true
	checksum, err := GetChecksum(app, appAsset.Checksum, release)
	if err != nil {
		if err == ErrNoChecksum {
			verify = false
		} else {
			return "", err
		}
	}
	// if there is a checksum, verify that the app is not the same as the internet using the hashes
	// if there is no checksum then the cache verification will be performed later on the function
	if verify && app.UseCache && cacheWithChecksum(checksum, app) {
		return "", ErrIsChached
	}

	client := NewGithubClient(app, nil)
	rc, lenght, err := downloadableAsset(client, *asset.URL)
	if err != nil {
		return "", err
	}

	tempPath := filepath.Join(Config().BasePath, *asset.Name)

	return CreateFile(rc, lenght, tempPath)
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

func verifyChecksum(checksum []byte, path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	verify, err := VerifyWithChecksum(checksum, f, NewFileHash()) // TODO make the hash algorithm be configurable
	if err != nil {
		return false
	}
	if !verify {
		return false
	}
	return true
}

func hashFile(path string, hasher hash.Hash) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

func cacheWithChecksum(checksum []byte, app configuration.Application) (isCached bool) {
	_mutexHandleAssetMatch.Lock()
	defer _mutexHandleAssetMatch.Unlock()

	file, err := os.Open(app.SystemPath)
	if err != nil {
		return
	}
	defer file.Close()

	hash, err := hashFile(app.SystemPath, NewFileHash())
	if err != nil {
		return
	}

	return slices.Equal(hash, checksum)
}

func cacheWithFile(path string, app configuration.Application) (isCached bool) {
	_mutexHandleAssetMatch.Lock()
	defer _mutexHandleAssetMatch.Unlock()

	hashFileDownload, err := hashFile(path, NewFileHash())
	if err != nil {
		return
	}

	hashApp, err := hashFile(app.SystemPath, NewFileHash())
	if err != nil {
		return
	}

	return slices.Equal(hashApp, hashFileDownload)
}
