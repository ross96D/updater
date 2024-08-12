//go:build windows

package taskservice

import (
	"github.com/ross96D/taskmaster"
	"github.com/rs/zerolog/log"
)

type TaskService struct {
	service taskmaster.TaskService
}

func New() (*TaskService, error) {
	ts, err := taskmaster.Connect()
	if err != nil {
		return nil, err
	}
	taskService := new(TaskService)
	taskService.service = ts
	return taskService, nil
}

func (ts *TaskService) Disconnect() {
	ts.service.Disconnect()
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

func (ts *TaskService) Run(path string) error {
	task, err := ts.service.GetRegisteredTask(path)
	if err != nil {
		return err
	}
	_, err = task.Run()
	return err
}

var ts *TaskService

func get() (*TaskService, error) {
	var err error
	if ts == nil {
		ts, err = New()
	}
	if ts != nil && !ts.service.IsConnected() {
		ts = nil
		return get()
	}
	return ts, err
}

func close() {
	ts.Disconnect()
	ts = nil
}

func Stop(path string) error {
	if path == "" {
		log.Warn().Msg("task path is empty. No op")
		return nil
	}

	ts, err := get()
	if err != nil {
		return err
	}
	return ts.Stop(path)
}

func Start(path string) error {
	if path == "" {
		log.Warn().Msg("task path is empty. No op")
		return nil
	}

	ts, err := get()
	if err != nil {
		return err
	}
	return ts.Run(path)
}
