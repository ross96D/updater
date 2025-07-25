package match

import (
	"io"
	"os"
	"os/exec"

	"github.com/ross96D/updater/share/configuration"
	"github.com/ross96D/updater/share/utils"
	taskservice "github.com/ross96D/updater/task_service"
	"github.com/rs/zerolog"
)

type IO interface {
	RunCommand(*zerolog.Logger, configuration.Command) error
	Unzip(string) error
	ServiceStart(string, taskservice.ServiceType) error
	ServiceStop(string, taskservice.ServiceType) error
	CopyFromReader(io.Reader, string) error
	RenameSafe(string, string) error
	Remove(string) error
	CreateCronjobConfiguration(serviceName string, jobs []cronJob) error
}

type implIO struct{}

func (implIO) RunCommand(logger *zerolog.Logger, command configuration.Command) error {
	return RunCommand(logger, command)
}

func (implIO) Unzip(path string) error {
	return utils.Unzip(path)
}

func (i implIO) ServiceStart(name string, st taskservice.ServiceType) error {
	return taskservice.NewService(st).Start(name)
}

func (i implIO) ServiceStop(name string, st taskservice.ServiceType) error {
	return taskservice.NewService(st).Stop(name)
}

func (implIO) CopyFromReader(reader io.Reader, dst string) error {
	return utils.CopyFromReader(reader, dst)
}

func (implIO) RenameSafe(oldpath string, newpath string) error {
	return utils.RenameSafe(oldpath, newpath)
}

func (implIO) Remove(path string) error {
	return os.Remove(path)
}

func (implIO) CreateCronjobConfiguration(serviceName string, jobs []cronJob) error {
	return createCronjobConfiguration(serviceName, jobs)
}

type dryRunIO struct{}

func (dryRunIO) RunCommand(logger *zerolog.Logger, command configuration.Command) error {
	cmd := exec.Command(command.Command, command.Args...)
	if command.Path != "" {
		cmd.Path = command.Path
	}
	logger.Info().Msg("running post command " + cmd.String())
	logger.Info().Msg("success post command " + cmd.String())

	return nil
}

func (dryRunIO) Unzip(_ string) error {
	return nil
}

func (dryRunIO) ServiceStart(_ string, _ taskservice.ServiceType) error {
	return nil
}

func (dryRunIO) ServiceStop(_ string, _ taskservice.ServiceType) error {
	return nil
}

func (dryRunIO) CopyFromReader(_ io.Reader, _ string) error {
	return nil
}

func (dryRunIO) RenameSafe(_ string, _ string) error {
	return nil
}

func (dryRunIO) Remove(_ string) error {
	return nil
}

func (dryRunIO) CreateCronjobConfiguration(serviceName string, jobs []cronJob) error {
	return nil
}
