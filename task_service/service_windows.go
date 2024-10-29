//go:build windows

package taskservice

import (
	"os/exec"

	"github.com/ross96D/taskmaster"
	"github.com/rs/zerolog/log"
)

type TaskService struct {
	service taskmaster.TaskService
}

var ts *TaskService

func newTS() (*TaskService, error) {
	ts, err := taskmaster.Connect()
	if err != nil {
		return nil, err
	}
	taskService := new(TaskService)
	taskService.service = ts
	return taskService, nil
}

func (ts *TaskService) Stop(path string) error {
	if path == "" {
		log.Info().Msg("task path is empty. No op")
		return nil
	}

	task, err := ts.service.GetRegisteredTask(path)
	if err != nil {
		return err
	}
	return task.Stop()
}

func (ts *TaskService) Start(path string) error {
	task, err := ts.service.GetRegisteredTask(path)
	if err != nil {
		return err
	}
	_, err = task.Run()
	return err
}

func get() (*TaskService, error) {
	var err error
	if ts == nil {
		ts, err = newTS()
	}
	if ts != nil && !ts.service.IsConnected() {
		ts = nil
		return get()
	}
	return ts, err
}

type NNSMService struct{}

func (NNSMService) Stop(name string) error {
	return exec.Command("nssm.exe", "stop", name).Run()
}

func (NNSMService) Start(name string) error {
	return exec.Command("nssm.exe", "start", name).Run()
}

func NewService(service ServiceType) Service {
	if service == NNSM {
		return NNSMService{}
	} else {
		resp, _ := get()
		return resp
	}
}
