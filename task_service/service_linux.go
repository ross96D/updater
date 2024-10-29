//go:build linux

package taskservice

import (
	"fmt"
	"os/exec"
)

type SystemctlService struct{}

func (ts SystemctlService) Stop(name string) error {
	cmd := exec.Command("systemctl", "stop", name)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("systemctl stop %s %w", string(out), err)
	}
	return nil
}

func (ts SystemctlService) Start(name string) error {
	cmd := exec.Command("systemctl", "start", name)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("systemctl start %s %w", string(out), err)
	}
	return nil
}

func NewService(service ServiceType) Service {
	return SystemctlService{}
}
