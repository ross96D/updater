package share

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/ross96D/updater/logger"
	"github.com/ross96D/updater/share/configuration"
	"github.com/ross96D/updater/share/utils"
	taskservice "github.com/ross96D/updater/task_service"
	"github.com/rs/zerolog"
)

type Data interface {
	Get(name string) io.ReadCloser
}

type NoData struct{}

func (NoData) Get(name string) io.ReadCloser { return nil }

type UpdateOpts func(*appUpdater)

func WithDryRun() UpdateOpts {
	return func(au *appUpdater) {
		au.dryRun = true
	}
}

func WithData(data Data) UpdateOpts {
	return func(au *appUpdater) {
		au.data = data
	}
}

func Update(ctx context.Context, app configuration.Application, opts ...UpdateOpts) error {
	u := NewAppUpdater(ctx, app, opts...)
	// TODO log on additional asset fail but do not return 500
	err := u.UpdateAdditionalAssets()
	err2 := u.UpdateTaskAssets()
	err3 := u.RunPostAction()
	return errors.Join(err, err2, err3)
}

type state int

const (
	correct = iota
	failed
)

type appUpdater struct {
	app  configuration.Application
	log  *zerolog.Logger
	data Data

	// how do i know if i am on a failed state?
	// if update aditional assets fail is a failed state?
	state  state
	dryRun bool
}

func (u appUpdater) seek(asset configuration.Asset) io.ReadCloser {
	return u.data.Get(asset.Name)
}

func NewAppUpdater(ctx context.Context, app configuration.Application, opts ...UpdateOpts) *appUpdater {
	appUpd := &appUpdater{
		app:   app,
		state: correct,
		log:   logger.ResponseWithLogger.FromContext(ctx),
	}

	for _, opt := range opts {
		opt(appUpd)
	}

	return appUpd
}

func (u *appUpdater) UpdateTaskAssets() error {
	var errs []error = make([]error, 0)

	for _, v := range u.app.Assets {
		// if v.ServicePath == "" then is not a Task Asset
		if v.ServicePath == "" {
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
		// if v.ServicePath != "" then is not an Additional Asset
		if v.ServicePath != "" {
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

func (u *appUpdater) updateTask(asset configuration.Asset) (err error) {
	var fnCopy func() error
	if fnCopy, err = u.updateAsset(asset); err != nil {
		return fmt.Errorf("updateTask %w", err)
	}

	// TODO this needs a mutex?
	u.log.Info().Msgf("stop %s", asset.ServicePath)
	if err = taskservice.Stop(asset.ServicePath); err != nil {
		u.log.Error().Err(err).Msgf("error stoping %s", asset.ServicePath)
		return fmt.Errorf("updateTask Stop() %w", err)
	}

	defer func() {
		u.log.Info().Msgf("start %s", asset.ServicePath)
		if err := taskservice.Start(asset.ServicePath); err != nil {
			// TODO Should i fail here?
			u.log.Error().Err(err).Msgf("error starting %s", asset.ServicePath)
		}
	}()

	return fnCopy()
}

func (u *appUpdater) updateAsset(asset configuration.Asset) (fnCopy func() (err error), err error) {
	data := u.seek(asset)
	if data == nil {
		msg := "updateAsset() no match " + asset.Name
		u.log.Warn().Msg(msg)
		return nil, errors.New(msg)
	}

	fnCopy = func() (err error) {
		defer data.Close()
		if err = utils.RenameSafe(asset.SystemPath, asset.SystemPath+".old"); err != nil {
			return
		}

		if err = utils.CopyFromReader(data, asset.SystemPath); err != nil {
			u.log.Error().
				Err(err).
				Msgf("Copying from %s. Rollback rename %s to %s",
					asset.Name, asset.SystemPath+".old", asset.SystemPath)
			os.Remove(asset.SystemPath)
			err2 := utils.RenameSafe(asset.SystemPath+".old", asset.SystemPath)
			if err2 != nil {
				u.log.Error().Err(err2).Msgf("rename fail %s to %s", asset.SystemPath+".old", asset.SystemPath)
			}
			return err
		}
		if asset.Unzip {
			u.log.Info().Msg("unzip: " + asset.SystemPath)
			if err = utils.Unzip(asset.SystemPath); err != nil {
				u.log.Error().Err(err).Msg("unzip: " + asset.SystemPath)
				return err
			}
		}

		if asset.Command != nil {
			cmd := exec.Command(asset.Command.Command, asset.Command.Args...)
			if asset.Command.Path != "" {
				cmd.Dir = asset.Command.Path
			}
			u.log.Info().Msg("Running command for " + asset.Name + ": " + cmd.String())
			// TODO this should pipe the data into the logger
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				u.log.Error().Err(err).Msgf("%s cmd: %s", asset.Name, cmd.String())
				// TODO should i leave the error be like this if i already kinda handle it?
				return fmt.Errorf("%s cmd: %s %w", asset.Name, cmd.String(), err)
			}
			u.log.Info().Msgf("success %s cmd: %s", asset.Name, cmd.String())
		}
		u.log.Info().Msgf("Asset %s update success", asset.Name)
		return nil
	}

	return
}

func (u *appUpdater) RunPostAction() error {
	if u.app.Command == nil || u.state == failed {
		return nil
	}
	cmd := exec.Command(u.app.Command.Command, u.app.Command.Args...)
	if u.app.Command.Path != "" {
		cmd.Dir = u.app.Command.Path
	}
	u.log.Info().Msg("running post command " + cmd.String())
	// TODO this should pipe the data into the logger
	err := cmd.Run()
	if err != nil {
		u.log.Error().Err(err).Msgf("post command %s", cmd.String())
		return fmt.Errorf("post command cmd: %s %w", cmd.String(), err)
	}
	u.log.Info().Msgf("success post command %s", cmd.String())
	return nil
}
