//go:build windows

package taskservice

import "github.com/ross96D/taskmaster"

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

func (ts *TaskService) GetRunningTasks() (taskmaster.RunningTaskCollection, error) {
	return ts.service.GetRunningTasks()
}

func (ts *TaskService) GetRegisteredTasks() (taskmaster.RegisteredTaskCollection, error) {
	return ts.service.GetRegisteredTasks()
}

func (ts *TaskService) Stop(path string) {
	task, err := ts.service.GetRegisteredTask(path)
	if err != nil {
		panic(err)
	}
	task.Stop()
}

func (ts *TaskService) Run(path string) {
	task, err := ts.service.GetRegisteredTask(path)
	if err != nil {
		panic(err)
	}
	task.Run()
}
