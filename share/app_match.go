package share

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/go-github/v60/github"
	"github.com/ross96D/updater/share/configuration"
	taskservice "github.com/ross96D/updater/task_service"
	"github.com/rs/zerolog/log"
)

func UpdateApp(app configuration.Application, release *github.RepositoryRelease) error {
	u := newAppUpdater(app, release)
	// TODO append errors
	err := u.UpdateAdditionalAssets()
	// TODO append errors
	err2 := u.UpdateTaskAssets()
	// TODO append errors
	err3 := u.RunPostAction()
	u.CleanUp()
	return errors.Join(err, err2, err3)
}

type appUpdater struct {
	app     configuration.Application
	release *github.RepositoryRelease

	cleanupFuncs []func()
}

func (u *appUpdater) addCleanUpFn(fn func()) {
	if u.cleanupFuncs == nil {
		u.cleanupFuncs = make([]func(), 0)
	}
	u.cleanupFuncs = append(u.cleanupFuncs, fn)
}

func (u appUpdater) seek(asset configuration.Asset) *github.ReleaseAsset {
	for _, githubAsset := range u.release.Assets {
		if *githubAsset.Name == asset.GetAsset() {
			return githubAsset
		}
	}
	return nil
}

func newAppUpdater(app configuration.Application, release *github.RepositoryRelease) *appUpdater {
	return &appUpdater{
		app:     app,
		release: release,
	}
}

func (u *appUpdater) UpdateTaskAssets() error {
	var errs []error = make([]error, 0)
	for _, v := range u.app.TaskAssets {
		if err := u.updateTask(v); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Join(errs...)
}

func (u *appUpdater) UpdateAdditionalAssets() error {
	var errs []error = make([]error, 0)
	for _, v := range u.app.AdditionalAssets {
		fnCopy, err := u.updateAsset(v)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if err = fnCopy(); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errors.Join(errs...)
}

func (u *appUpdater) updateTask(v configuration.TaskAsset) (err error) {
	var fnCopy func() error
	if fnCopy, err = u.updateAsset(v); err != nil {
		return
	}

	// TODO this needs a mutex
	if err = taskservice.Stop(v.TaskSchedPath); err != nil {
		return
	}

	defer func() {
		if err := taskservice.Start(v.TaskSchedPath); err != nil {
			log.Error().Err(fmt.Errorf("reruning the task %s %w", v.TaskSchedPath, err)).Send()
		}
	}()

	return fnCopy()
}

func (u *appUpdater) updateAsset(v configuration.Asset) (fnCopy func() (err error), err error) {
	releaseAsset := u.seek(v)
	if releaseAsset == nil {
		err = fmt.Errorf("no match for asset %s", v.GetAsset())
		return
	}

	var verif verifier = func(io.Reader) bool { return true }

	if u.app.UseCache {
		verif, err = checksumVerifier(NewGetChecksum(NewGithubClient(u.app, nil), u.app, v.GetChecksum(), v, u.release))
		if err != nil {
			return
		}
	}
	rc, lenght, err := downloadableAsset(NewGithubClient(u.app, nil), *releaseAsset.URL)
	if err != nil {
		return
	}

	downloadTempPath := filepath.Join(Config().BasePath, *releaseAsset.Name)
	log.Info().Msg("save asset to temporary file in " + downloadTempPath)
	if downloadTempPath, err = CreateFile(rc, lenght, downloadTempPath); err != nil {
		return
	}

	u.addCleanUpFn(func() {
		log.Info().Msgf("removing temporary file %s", downloadTempPath)
		os.Remove(downloadTempPath)
	})

	f, err := os.Open(downloadTempPath)
	if err != nil {
		return
	}
	if !verif(f) {
		f.Close()
		log.Warn().Msgf("asset %s could not be verified", v.GetAsset())
		err = ErrUnverifiedAsset
		return
	}
	f.Close()

	fnCopy = func() (err error) {
		if err = RenameSafe(v.GetSystemPath(), v.GetSystemPath()+".old"); err != nil {
			return
		}
		if err = Copy(downloadTempPath, v.GetSystemPath()); err != nil {
			os.Remove(v.GetSystemPath())
			log.Warn().Msgf("roll back rename %s to %s", v.GetSystemPath()+".old", v.GetSystemPath())
			_ = RenameSafe(v.GetSystemPath()+".old", v.GetSystemPath())
			return nil
		}
		return nil
	}

	return
}

func (u appUpdater) RunPostAction() error {
	if u.app.PostAction == nil {
		return nil
	}
	cmd := exec.Command(u.app.PostAction.Command, u.app.PostAction.Args...)
	log.Info().Msg("running post action " + cmd.String())
	b, err := cmd.Output()
	if err != nil {
		log.Error().Err(fmt.Errorf(
			"running post action %s with output %s. Error: %w",
			cmd.String(),
			string(b),
			err,
		)).Send()
		return err
	}
	log.Info().Str("command output", string(b)).Send()
	return nil
}

type verifier func(io.Reader) bool

func checksumVerifier(gchsm IGetChecksum) (verifier, error) {
	checksum, err := gchsm.GetChecksum()
	if err != nil && err != ErrNoChecksum {
		return nil, err
	}

	var verif verifier = func(r io.Reader) bool {
		if r == nil {
			// TODO add context
			log.Error().Err(fmt.Errorf("nil reader on verifier")).Send()
			return false
		}

		h := NewHasher()
		_, err := io.Copy(h, r)
		if err != nil {
			// TODO add context
			log.Error().Err(fmt.Errorf("error hashing %w", err)).Send()
			return false
		}
		hashed := h.Sum(nil)
		return bytes.Equal(hashed, checksum)
	}

	return verif, nil
}

func (u appUpdater) CleanUp() {
	for _, f := range u.cleanupFuncs {
		f()
	}
}
