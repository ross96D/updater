//go:build !windows

package taskservice

import (
	"fmt"
	"os/exec"
)

type TaskService struct {
}

func New() (*TaskService, error) {
	return &TaskService{}, nil
}

func (ts *TaskService) Disconnect() {}

func (ts *TaskService) Stop(name string) error {
	cmd := exec.Command("systemctl", "stop", name)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("systemctl stop %s %w", string(out), err)
	}
	return nil
}

func (ts *TaskService) Run(name string) error {
	cmd := exec.Command("systemctl", "start", name)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("systemctl start %s %w", string(out), err)
	}
	return nil
}

func Stop(name string) error {
	ts, err := New()
	if err != nil {
		return err
	}
	return ts.Stop(name)
}

func Start(name string) error {
	ts, err := New()
	if err != nil {
		return err
	}
	return ts.Run(name)
}
