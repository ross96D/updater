package match

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/ross96D/updater/logger"
	"github.com/ross96D/updater/share/configuration"
	taskservice "github.com/ross96D/updater/task_service"
	"github.com/rs/zerolog"
)

type Data interface {
	Get(name string) io.ReadCloser
	Clean()
}

type NoData struct{}

func (NoData) Get(name string) io.ReadCloser { return nil }
func (NoData) Clean()                        {}

type EmptyData struct{}

func (EmptyData) Get(name string) io.ReadCloser { return io.NopCloser(bytes.NewReader([]byte{})) }
func (EmptyData) Clean()                        {}

type UpdateOpts func(*appUpdater)

func WithDryRun(dryRun bool) UpdateOpts {
	return func(au *appUpdater) {
		if dryRun {
			au.io = dryRunIO{}
		}
	}
}

func WithData(data Data) UpdateOpts {
	return func(au *appUpdater) {
		au.data = data
	}
}

func Update(ctx context.Context, app configuration.Application, opts ...UpdateOpts) (errs JoinErrors) {
	u := NewAppUpdater(ctx, app, opts...)
	defer u.data.Clean()

	if u.app.Service != "" {
		u.log.Info().Msgf("stoping app level service %s", u.app.Service)
		err := u.io.ServiceStop(u.app.Service, taskservice.ServiceTypeFrom(app.ServiceType))
		errs.Add(err)
		defer func() {
			u.log.Info().Msgf("starting app level service %s", u.app.Service)
			errServiceStart := u.io.ServiceStart(u.app.Service, taskservice.ServiceTypeFrom(app.ServiceType))
			errs.Add(errServiceStart)
		}()
	}

	jobs, err := ValidateCronJobConfiguration(u.getJobContent())
	if err != nil {
		errs.Add(err)
		return
	}

	err = u.RunPreAction()
	errs.Add(err)

	err2 := u.UpdateAssets()
	errs.Concat(err2)

	err = u.RunPostAction()
	errs.Add(err)

	if !errs.LevelIsError() {
		u.io.CreateCronjobConfiguration(app.Name, jobs)
	}

	return
}

type appUpdater struct {
	app  configuration.Application
	log  *zerolog.Logger
	data Data
	io   IO
}

func (u appUpdater) getJobContent() []byte {
	jobReader := u.data.Get("__jobs")
	if jobReader == nil {
		return []byte{}
	}
	data, err := io.ReadAll(jobReader)
	if err != nil {
		return []byte{}
	}
	return data
}

func (u appUpdater) seek(asset configuration.Asset) io.ReadCloser {
	return u.data.Get(asset.Name)
}

func NewAppUpdater(ctx context.Context, app configuration.Application, opts ...UpdateOpts) *appUpdater {
	l, _ := logger.LoggerCtx_FromContext(ctx)
	appUpd := &appUpdater{
		app: app,
		log: l,
		io:  implIO{},
	}

	for _, opt := range opts {
		opt(appUpd)
	}

	return appUpd
}

func (u *appUpdater) UpdateAssets() (errs JoinErrors) {
	mut := &sync.Mutex{}
	append_errors := func(joinerr *JoinErrors, err ...error) {
		mut.Lock()
		if joinerr != nil {
			errs.Concat(*joinerr)
		}
		for _, e := range err {
			errs.Add(e)
		}
		mut.Unlock()
	}

	wg := &sync.WaitGroup{}
	var updateAsset = func(logger zerolog.Logger, asset configuration.Asset, wg *sync.WaitGroup) {
		if asset.Service != "" {
			err := u.updateTask(logger, asset)
			append_errors(&err)
		} else {
			var fnCopy func() (err error)
			var err error
			if fnCopy, err = u.updateAsset(logger, asset); err != nil {
				append_errors(nil, err)
			} else {
				if err = fnCopy(); err != nil {
					append_errors(nil, err)
				}
			}
		}
		wg.Done()
	}
	var i = 0
	// executes independent assets  concurrently
	for ; i < len(u.app.AsstesOrder); i++ {
		asset := u.app.AsstesOrder[i]
		if !asset.Independent {
			break
		}
		wg.Add(1)
		assetLogger := u.log.With().Logger()
		assetLogger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("asset", asset.Name)
		})
		go updateAsset(assetLogger, asset.Asset, wg)
	}

	wg.Wait()

	for ; i < len(u.app.AsstesOrder); i++ {
		asset := u.app.AsstesOrder[i]
		assetLogger := u.log.With().Logger()
		assetLogger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("asset", asset.Name)
		})
		if asset.Service != "" {
			err := u.updateTask(assetLogger, asset.Asset)
			errs.Concat(err)
		} else {
			var fnCopy func() (err error)
			var err error
			if fnCopy, err = u.updateAsset(assetLogger, asset.Asset); err != nil {
				errs.Add(err)
			} else {
				if err = fnCopy(); err != nil {
					errs.Add(err)
				}
			}
		}
	}
	return
}

func (u *appUpdater) updateTask(logger zerolog.Logger, asset configuration.Asset) (errs JoinErrors) {
	var fnCopy func() error
	var err error
	if fnCopy, err = u.updateAsset(logger, asset); err != nil {
		errs.Add(FmtFromInnerError("updateTask %w", err))
		return
	}

	// TODO this needs a mutex?
	logger.Info().Msgf("stop %s", asset.Service)
	if err = u.io.ServiceStop(asset.Service, taskservice.ServiceTypeFrom(asset.ServiceType)); err != nil {
		logger.Warn().Err(err).Msgf("error stoping %s", asset.Service)
		errs.Add(ErrWarning{fmt.Errorf("updateTask Stop() %w", err)})
	}

	defer func() {
		logger.Info().Msgf("start %s", asset.Service)
		if err := u.io.ServiceStart(asset.Service, taskservice.ServiceTypeFrom(asset.ServiceType)); err != nil {
			logger.Warn().Err(err).Msgf("error starting %s", asset.Service)
			errs.Add(ErrError{err})
		}
	}()

	err = fnCopy()
	errs.Add(err)
	return
}

func (u *appUpdater) updateAsset(logger zerolog.Logger, asset configuration.Asset) (fnCopy func() (err error), err error) {
	data := u.seek(asset)
	if data == nil {
		msg := "updateAsset() no match " + asset.Name
		logger.Warn().Msg(msg)
		return nil, ErrWarning{errors.New(msg)}
	}

	fnCopy = func() (err error) {
		defer data.Close()

		if asset.CommandPre != nil {
			logger := logger.With().Logger()
			logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str("asset", asset.Name).Str("kind", "pre")
			})
			logger.Info().Msg("Running pre action commnad")
			err = u.io.RunCommand(&logger, *asset.CommandPre)
			logger.Info().Msg("Finished running pre action commnad")
			if err != nil {
				return err
			}
		}

		SystemPathOld := asset.SystemPath + ".old"

		if err = u.io.RenameSafe(asset.SystemPath, SystemPathOld); err != nil {
			return ErrError{err}
		}

		rollback := func() {
			u.io.Remove(asset.SystemPath) //nolint: errcheck
			err2 := u.io.RenameSafe(SystemPathOld, asset.SystemPath)
			if err2 != nil {
				logger.Error().Err(err2).Msgf("move fail %s to %s", SystemPathOld, asset.SystemPath)
			}
		}

		logger.Info().Msgf("Copying from %s to %s", asset.Name, asset.SystemPath)
		if err = u.io.CopyFromReader(data, asset.SystemPath); err != nil {
			logger.Error().Err(err).Msgf("Copying from %s to %s. Rollback, move %s to %s", asset.Name, asset.SystemPath, SystemPathOld, asset.SystemPath)
			rollback()
			return ErrError{err}
		}

		if asset.Unzip {
			logger.Info().Msg("unzip: " + asset.SystemPath)
			if err = u.io.Unzip(asset.SystemPath); err != nil {
				logger.Error().Err(err).Msg("unzip: " + asset.SystemPath)
				rollback()
				return ErrError{err}
			}
		}

		if asset.Command != nil {
			logger := logger.With().Logger()
			logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str("asset", asset.Name).Str("kind", "post")
			})
			logger.Info().Msg("Running post action command")
			err = u.io.RunCommand(&logger, *asset.Command)
			logger.Info().Msg("Finished running post action command")
			if err != nil {
				return err
			}
		}

		if !asset.KeepOld {
			u.io.Remove(SystemPathOld) //nolint: errcheck
		}
		logger.Info().Msgf("Asset %s updated successfully", asset.Name)
		return nil
	}
	return
}

func (u *appUpdater) RunPreAction() error {
	if u.app.CommandPre == nil {
		return nil
	}
	u.log.Info().Msg("Running pre action command")
	err := u.io.RunCommand(u.log, *u.app.CommandPre)
	u.log.Info().Msg("Finish running pre action command")
	return err
}

func (u *appUpdater) RunPostAction() error {
	if u.app.Command == nil {
		return nil
	}
	u.log.Info().Msg("Running post action command")
	err := u.io.RunCommand(u.log, *u.app.Command)
	u.log.Info().Msg("Finish running post action command")
	return err
}
