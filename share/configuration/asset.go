package configuration

type Asset struct {
	Name        string   `json:"name"`
	SystemPath  string   `json:"system_path"`
	Service     string   `json:"service"`
	ServiceType string   `json:"service_type"`
	KeepOld     bool     `json:"keep_old"`
	Unzip       bool     `json:"unzip"`
	CommandPre  *Command `json:"cmd_pre"`
	Command     *Command `json:"cmd"`
}

type AssetOrder struct {
	Asset
	Independent bool
}
