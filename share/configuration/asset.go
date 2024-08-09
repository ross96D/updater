package configuration

type Asset struct {
	Name          string `json:"name"`
	SystemPath    string `json:"system_path"`
	TaskSchedPath string `json:"task_sched_path"`
	Unzip         bool   `json:"unzip"`
}
