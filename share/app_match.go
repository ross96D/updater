package share

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/ross96D/updater/share/configuration"
	"github.com/ross96D/updater/share/utils"
	taskservice "github.com/ross96D/updater/task_service"
	"github.com/rs/zerolog/log"
)

type Data interface {
	Get(name string) io.ReadCloser
}

type NoData struct{}

func (NoData) Get(name string) io.ReadCloser { return nil }

func Update(app configuration.Application, data Data) error {
	u := NewAppUpdater(app, data)
	// TODO log on additional asset fail but do not return 500
	err := u.UpdateAdditionalAssets()
	err2 := u.UpdateTaskAssets()
	err3 := u.RunPostAction()
	// u.CleanUp()
	return errors.Join(err, err2, err3)
}

type state int

const (
	correct = iota
	failed
)

type appUpdater struct {
	app  configuration.Application
	data Data

	// how do i know if i am on a failed state?
	// if update aditional assets fail is a failed state?
	state state

	// cleanupFuncs []func()
}

// func (u *appUpdater) addCleanUpFn(fn func()) {
// 	if u.cleanupFuncs == nil {
// 		u.cleanupFuncs = make([]func(), 0)
// 	}
// 	u.cleanupFuncs = append(u.cleanupFuncs, fn)
// }

func (u appUpdater) seek(asset configuration.Asset) io.ReadCloser {
	return u.data.Get(asset.Name)
}

func NewAppUpdater(app configuration.Application, data Data) *appUpdater {
	return &appUpdater{
		app:   app,
		data:  data,
		state: correct,
	}
}

func (u *appUpdater) UpdateTaskAssets() error {
	var errs []error = make([]error, 0)

	for _, v := range u.app.Assets {
		if v.TaskSchedPath == "" {
			continue
		}

		if err := u.updateTask(v); err != nil {
			errs = append(errs, err)
			continue
		}
	}
	err := errors.Join(errs...)
	if err != nil {
		u.state = failed
	}
	return err
}

func (u *appUpdater) UpdateAdditionalAssets() error {
	var errs []error = make([]error, 0)
	for _, v := range u.app.Assets {
		if v.TaskSchedPath != "" {
			continue
		}
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

func (u *appUpdater) updateTask(v configuration.Asset) (err error) {
	var fnCopy func() error
	if fnCopy, err = u.updateAsset(v); err != nil {
		return fmt.Errorf("updateTask updateAsset() %w", err)
	}

	// TODO this needs a mutex?
	if err = taskservice.Stop(v.TaskSchedPath); err != nil {
		return fmt.Errorf("updateTask Stop() %w", err)
	}

	defer func() {
		if err := taskservice.Start(v.TaskSchedPath); err != nil {
			log.Error().Err(fmt.Errorf("reruning the task %s %w", v.TaskSchedPath, err)).Send()
		}
	}()

	return fnCopy()
}

func (u *appUpdater) updateAsset(v configuration.Asset) (fnCopy func() (err error), err error) {
	assetData := u.seek(v)
	if assetData == nil {
		err = fmt.Errorf("no match for asset %s", v.Name)
		return
	}

	fnCopy = func() (err error) {
		defer assetData.Close()
		if err = utils.RenameSafe(v.SystemPath, v.SystemPath+".old"); err != nil {
			return
		}

		if err = utils.CopyFromReader(assetData, v.SystemPath); err != nil {
			os.Remove(v.SystemPath)
			log.Debug().Msgf("roll back rename %s to %s", v.SystemPath+".old", v.SystemPath)
			_ = utils.RenameSafe(v.SystemPath+".old", v.SystemPath)
			return nil
		}
		if v.Unzip {
			if err = utils.Unzip(v.SystemPath); err != nil {
				return err
			}
		}

		return nil
	}

	return
}

func (u appUpdater) RunPostAction() error {
	if u.app.PostAction == nil || u.state == failed {
		return nil
	}
	cmd := exec.Command(u.app.PostAction.Command, u.app.PostAction.Args...)
	log.Info().Msg("running post action " + cmd.String())
	b, err := cmd.Output()
	if err != nil {
		log.Error().Err(
			fmt.Errorf(
				"running post action %s with output %s. Error: %w",
				cmd.String(),
				string(b),
				err,
			),
		).Send()
		return err
	}
	log.Info().Str("command output", string(b)).Send()
	return nil
}

// func (u appUpdater) CleanUp() {
// 	for _, f := range u.cleanupFuncs {
// 		f()
// 	}
// }
