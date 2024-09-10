package match

import (
	"io"
	"os"

	"github.com/ross96D/updater/share/utils"
	taskservice "github.com/ross96D/updater/task_service"
)

type IO interface {
	RunCommand(string, ...string) error
	Unzip(string) error
	ServiceStart(string) error
	ServiceStop(string) error
	CopyFromReader(io.Reader, string) error
	RenameSafe(string, string) error
	Remove(string) error
}

type implIO struct{}

func (implIO) RunCommand(name string, args ...string) error {
	panic("not implemented") // TODO: Implement
}

func (implIO) Unzip(path string) error {
	return utils.Unzip(path)
}

func (implIO) ServiceStart(name string) error {
	return taskservice.Start(name)
}

func (implIO) ServiceStop(name string) error {
	return taskservice.Stop(name)
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

type dryRunIO struct{}

func (dryRunIO) RunCommand(_ string, _ ...string) error {
	return nil
}

func (dryRunIO) Unzip(_ string) error {
	return nil
}

func (dryRunIO) ServiceStart(_ string) error {
	return nil
}

func (dryRunIO) ServiceStop(_ string) error {
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
