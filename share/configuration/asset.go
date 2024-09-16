package configuration

type Asset struct {
	Name        string   `json:"name"`
	SystemPath  string   `json:"system_path"`
	ServicePath string   `json:"service"`
	KeepOld     bool     `json:"keep_old"`
	Unzip       bool     `json:"unzip"`
	Command     *Command `json:"cmd"`
}
