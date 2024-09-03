package configuration

type Asset struct {
	Name          string   `json:"name"`
	SystemPath    string   `json:"system_path"`
	TaskSchedPath string   `json:"service"`
	Unzip         bool     `json:"unzip"`
	Command       *Command `json:"cmd"`
}
