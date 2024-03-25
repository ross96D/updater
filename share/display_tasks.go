package share

import (
	"fmt"

	"github.com/ross96D/taskmaster"
)

func DisplayTasks() (err error) {
	service, err := taskmaster.Connect()
	if err != nil {
		return
	}
	tasks, err := service.GetRunningTasks()
	if err != nil {
		return
	}
	for _, t := range tasks {
		fmt.Printf("%+v\n", t)
	}
	return
}
