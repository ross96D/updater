package match

import (
	"context"
	"errors"
	"fmt"
	"io"
	"unsafe"

	"github.com/ross96D/updater/logger"
	"github.com/ross96D/updater/share/configuration"
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

type ErrError struct{ err error }
type ErrErrors struct{ errs []error }

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
		if _, ok := err.(ErrError); ok {
			isError = true
		}
		if _, ok := err.(ErrErrors); ok {
			isError = true
		}
		if err != nil {
			noNilErrs = append(noNilErrs, err)
		}
	}
	if isError {
		return ErrErrors{noNilErrs}
	}
	return errors.Join(noNilErrs...)
}

type Data interface {
	Get(name string) io.ReadCloser
}

type NoData struct{}

func (NoData) Get(name string) io.ReadCloser { return nil }

type UpdateOpts func(*appUpdater)

func WithDryRun() UpdateOpts {
	return func(au *appUpdater) {
		au.io = dryRunIO{}
	}
}

func WithData(data Data) UpdateOpts {
	return func(au *appUpdater) {
		au.data = data
	}
}

func Update(ctx context.Context, app configuration.Application, opts ...UpdateOpts) error {
	u := NewAppUpdater(ctx, app, opts...)
	err := u.UpdateAdditionalAssets()
	err2 := u.UpdateTaskAssets()
	err3 := u.RunPostAction()
	return PackError(err, err2, err3)
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
	appUpd := &appUpdater{
		app: app,
		log: logger.ResponseWithLogger.FromContext(ctx),
		io:  implIO{},
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
	return PackError(errs...)
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
	return PackError(errs...)
}

func (u *appUpdater) updateTask(asset configuration.Asset) (err error) {
	var fnCopy func() error
	if fnCopy, err = u.updateAsset(asset); err != nil {
		return fmt.Errorf("updateTask %w", err)
	}

	// TODO this needs a mutex?
	u.log.Info().Msgf("stop %s", asset.ServicePath)
	if err = u.io.ServiceStop(asset.ServicePath); err != nil {
		u.log.Error().Err(err).Msgf("error stoping %s", asset.ServicePath)
		return ErrError{fmt.Errorf("updateTask Stop() %w", err)}
	}

	defer func() {
		u.log.Info().Msgf("start %s", asset.ServicePath)
		if err := u.io.ServiceStart(asset.ServicePath); err != nil {
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
		if err = u.io.RenameSafe(asset.SystemPath, asset.SystemPath+".old"); err != nil {
			return ErrError{err}
		}

		if err = u.io.CopyFromReader(data, asset.SystemPath); err != nil {
			u.log.Error().Err(err).Msgf("Copying from %s. Rollback rename %s to %s", asset.Name, asset.SystemPath+".old", asset.SystemPath)

			_ = u.io.Remove(asset.SystemPath)
			err2 := u.io.RenameSafe(asset.SystemPath+".old", asset.SystemPath)
			if err2 != nil {
				u.log.Error().Err(err2).Msgf("rename fail %s to %s", asset.SystemPath+".old", asset.SystemPath)
			}
			return ErrError{err}
		}
		if asset.Unzip {
			u.log.Info().Msg("unzip: " + asset.SystemPath)
			if err = u.io.Unzip(asset.SystemPath); err != nil {
				u.log.Error().Err(err).Msg("unzip: " + asset.SystemPath)
				return ErrError{err}
			}
		}

		if asset.Command != nil {
			//! TODO
			logger := u.log.With().Logger()
			logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str("asset", asset.Name)
			})
			err = u.io.RunCommand(&logger, *asset.Command)
			if err != nil {
				return err
			}
		}
		u.log.Info().Msgf("Asset %s update success", asset.Name)
		return nil
	}

	return
}

func (u *appUpdater) RunPostAction() error {
	if u.app.Command == nil {
		return nil
	}
	err := u.io.RunCommand(u.log, *u.app.Command)
	return err
}