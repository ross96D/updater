package share

import (
	"context"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sync"

	"github.com/google/go-github/v60/github"
	"github.com/ross96D/updater/share/configuration"
	taskservice "github.com/ross96D/updater/task_service"
	"github.com/rs/zerolog/log"
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
	log.Info().Msg(fmt.Sprint("should verify the asset:", verify))

	// if there is a checksum, verify that the app is not the same as the one on the repo using the hashes
	// if there is no checksum then the cache verification will be performed later on the function
	if verify && app.UseCache && cacheWithChecksum(checksum, app) {
		return ErrIsChached
	}

	log.Info().Msg("Donwloading asset " + app.AssetName)
	rc, lenght, err := downloadableAsset(NewGithubClient(app, nil), *asset.URL)
	if err != nil {
		return err
	}

	tempPath := filepath.Join(Config().BasePath, *asset.Name)
	log.Info().Msg("save asset to temporary file in " + tempPath)
	if tempPath, err = CreateFile(rc, lenght, tempPath); err != nil {
		return err
	}
	// remove temp file
	defer func() {
		log.Info().Msg("removing temporary file")
		os.Remove(tempPath)
	}()

	if verify {
		log.Info().Msg("verifiying checksum")
		// verifiy that the checksum correspond to the downloaded asset
		if !verifyChecksum(checksum, tempPath) {
			return ErrUnverifiedAsset
		}
	} else {
		log.Info().Msg("checking if asset is already installed")
		// as there is no checksum hash the both, the app file and the downloaded one.
		if cacheWithFile(tempPath, app) {
			log.Info().Msg("do not return, but asset is cached")
			// return ErrIsChached
		}
	}

	additionalAssetsTempPath, err := CreateAdditionalAssets(app, release)
	if err != nil {
		log.Error().Err(fmt.Errorf("error downloading assets %w", err)).Send()
	}
	log.Info().Msg(fmt.Sprintf("additional paths are %d %+v", len(additionalAssetsTempPath), additionalAssetsTempPath))

	log.Info().Msg("stoping task " + app.TaskSchedPath)
	if err = taskservice.Stop(app.TaskSchedPath); err != nil {
		return err
	}
	defer func() {
		log.Info().Msg("Run the task " + app.TaskSchedPath)
		if err := taskservice.Start(app.TaskSchedPath); err != nil {
			log.Error().Err(fmt.Errorf("reruning the task %w", err)).Send()
		}
		if app.PostAction != nil {
			go func() {
				runPostAction(app)
			}()
		}
	}()

	_mutexHandleAssetMatch.Lock()
	defer _mutexHandleAssetMatch.Unlock()
	log.Info().Msg("Moving " + app.SystemPath + " to " + app.SystemPath + ".old")
	if err = RenameSafe(app.SystemPath, app.SystemPath+".old"); err != nil {
		return err
	}

	log.Info().Msg("Moving " + tempPath + " to " + app.SystemPath)
	if err = Copy(tempPath, app.SystemPath); err != nil {
		// Roll back
		log.Error().Err(fmt.Errorf("copy %s to %s %w", tempPath, app.SystemPath, err)).Send()
		os.Remove(app.SystemPath)
		RenameSafe(app.SystemPath+".old", app.SystemPath)
		return err
	}

	for _, p := range additionalAssetsTempPath {
		if err = RenameSafe(p.SystemPath, p.SystemPath+".old"); err != nil {
			// log error
			log.Warn().Err(fmt.Errorf("renameSafe %s to %s %w", p.SystemPath, p.SystemPath+".old", err)).Send()
			continue
		}
		if err = Copy(p.TempPath, p.SystemPath); err != nil {
			// Roll back
			log.Error().Err(fmt.Errorf("copy %s to %s %w", p.TempPath, p.SystemPath, err)).Send()
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
			log.Debug().Msg(a.Name + " not found")
			continue
		}
		log.Debug().Msg(a.Name + " was found")

		path, err := HandleAdditionalAsset(app, a, release.Assets[index], release)
		if err != nil {
			log.Warn().Err(fmt.Errorf("handling additional asset %s %w", a.Name, err)).Send()
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

func runPostAction(app configuration.Application) error {
	// TODO log the output of the command
	cmd := exec.Command(app.PostAction.Command, app.PostAction.Args...)
	log.Info().Msg("running post action " + cmd.String())
	b, err := cmd.Output()
	if err != nil {
		log.Error().Err(fmt.Errorf(
			"running post action %s %w",
			cmd.String(),
			err,
		)).Send()
		return err
	} else {
		log.Info().Str("command output", string(b))
	}
	return nil
}
