//go:build !windows

package taskservice

type TaskService struct {
	_ any
}

func New() (*TaskService, error) {
	panic("taskservice is a only windows service")
}

func (ts *TaskService) Disconnect() {
	panic("taskservice is a only windows service")
}

func (ts *TaskService) GetRunningTasks() (any, error) {
	panic("taskservice is a only windows service")
}

func (ts *TaskService) GetRegisteredTasks() (any, error) {
	panic("taskservice is a only windows service")
}

func (ts *TaskService) Stop(path string) {
	panic("taskservice is a only windows service")
}

func (ts *TaskService) Run(path string) {
	panic("taskservice is a only windows service")
}
