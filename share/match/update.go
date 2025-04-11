package match

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"unsafe"

	"github.com/ross96D/updater/logger"
	"github.com/ross96D/updater/share/configuration"
	taskservice "github.com/ross96D/updater/task_service"
	"github.com/rs/zerolog"
)

type ErrLevel interface {
	error
	Level() string
}

func joinErrorsMessage(errs []error) string {
	if len(errs) == 1 {
		return errs[0].Error()
	}

	b := []byte(errs[0].Error())
	for _, err := range errs[1:] {
		b = append(b, '\t')
		b = append(b, err.Error()...)
	}
	// At this point, b has at least one byte '\n'.
	return unsafe.String(&b[0], len(b))
}

// Error that indicates a fail in the update
type ErrError struct{ err error }

// A collection of errors that indicates fail in the update
type ErrErrors struct{ errs []error }

func NewErrError(err error) ErrError      { return ErrError{err: err} }
func NewErrErrors(errs []error) ErrErrors { return ErrErrors{errs: errs} }

func (e ErrErrors) Error() string { return joinErrorsMessage(e.errs) }
func (e ErrError) Error() string  { return e.err.Error() }
func (ErrErrors) Level() string   { return "error" }
func (ErrError) Level() string    { return "error" }

func PackError(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}

	isError := false
	noNilErrs := make([]error, 0, len(errs))
	for _, err := range errs {
		if err == nil {
			continue
		}
		if _, ok := err.(ErrError); ok {
			isError = true
		}
		if _, ok := err.(ErrErrors); ok {
			isError = true
		}
		noNilErrs = append(noNilErrs, err)
	}
	if len(noNilErrs) == 0 {
		return nil
	}

	if isError {
		return ErrErrors{noNilErrs}
	}
	return errors.Join(noNilErrs...)
}

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

func Update(ctx context.Context, app configuration.Application, opts ...UpdateOpts) (err error) {
	u := NewAppUpdater(ctx, app, opts...)
	defer u.data.Clean()

	if u.app.Service != "" {
		u.log.Info().Msgf("stoping app level service %s", u.app.Service)
		err = u.io.ServiceStop(u.app.Service, taskservice.ServiceTypeFrom(app.ServiceType))
		defer func() {
			u.log.Info().Msgf("starting app level service %s", u.app.Service)
			errServiceStart := u.io.ServiceStart(u.app.Service, taskservice.ServiceTypeFrom(app.ServiceType))
			err = PackError(errServiceStart, err)
		}()
	}
	err1 := u.RunPreAction()

	err2 := u.UpdateAssets()
	err3 := u.RunPostAction()
	err = PackError(err, err1, err2, err3)
	return
}

type appUpdater struct {
	app  configuration.Application
	log  *zerolog.Logger
	data Data
	io   IO
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

func (u *appUpdater) UpdateAssets() error {
	var errs []error = make([]error, 0)
	mut := &sync.Mutex{}
	append_errors := func(err error) {
		mut.Lock()
		errs = append(errs, err)
		mut.Unlock()
	}

	wg := &sync.WaitGroup{}
	var updateAsset = func(logger zerolog.Logger, asset configuration.Asset, wg *sync.WaitGroup) {
		if asset.Service != "" {
			if err := u.updateTask(logger, asset); err != nil {
				append_errors(err)
			}
		} else {
			var fnCopy func() (err error)
			var err error
			if fnCopy, err = u.updateAsset(logger, asset); err != nil {
				append_errors(err)
			} else {
				if err = fnCopy(); err != nil {
					append_errors(err)
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
			if err := u.updateTask(assetLogger, asset.Asset); err != nil {
				errs = append(errs, err)
			}
		} else {
			var fnCopy func() (err error)
			var err error
			if fnCopy, err = u.updateAsset(assetLogger, asset.Asset); err != nil {
				errs = append(errs, err)
			} else {
				if err = fnCopy(); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}
	return PackError(errs...)
}

func (u *appUpdater) updateTask(logger zerolog.Logger, asset configuration.Asset) (err error) {
	var fnCopy func() error
	if fnCopy, err = u.updateAsset(logger, asset); err != nil {
		return fmt.Errorf("updateTask %w", err)
	}

	// TODO this needs a mutex?
	logger.Info().Msgf("stop %s", asset.Service)
	if err = u.io.ServiceStop(asset.Service, taskservice.ServiceTypeFrom(asset.ServiceType)); err != nil {
		logger.Warn().Err(err).Msgf("error stoping %s", asset.Service)
		err = fmt.Errorf("updateTask Stop() %w", err)
	}

	defer func() {
		logger.Info().Msgf("start %s", asset.Service)
		if err1 := u.io.ServiceStart(asset.Service, taskservice.ServiceTypeFrom(asset.ServiceType)); err != nil {
			logger.Warn().Err(err1).Msgf("error starting %s", asset.Service)
			err = PackError(err, err1)
		}
	}()

	err1 := fnCopy()
	return PackError(err, err1)
}

func (u *appUpdater) updateAsset(logger zerolog.Logger, asset configuration.Asset) (fnCopy func() (err error), err error) {
	data := u.seek(asset)
	if data == nil {
		msg := "updateAsset() no match " + asset.Name
		logger.Warn().Msg(msg)
		return nil, errors.New(msg)
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
