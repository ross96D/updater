package configuration

type Asset interface {
	GetAsset() string
	GetChecksum() Checksum
	GetUnzip() bool
}

type AdditionalAsset struct {
	Name       string   `json:"name"`
	SystemPath string   `json:"system_path"`
	Checksum   Checksum `json:"checksum"`
	Unzip      bool     `json:"unzip"`
}

func (a AdditionalAsset) GetAsset() string {
	return a.Name
}

func (a AdditionalAsset) GetChecksum() Checksum {
	return a.Checksum
}

func (a AdditionalAsset) GetUnzip() bool {
	return a.Unzip
}

type TaskAsset struct {
	Name          string   `json:"name"`
	SystemPath    string   `json:"system_path"`
	TaskSchedPath string   `json:"task_sched_path"`
	Checksum      Checksum `json:"checksum"`
	Unzip         bool     `json:"unzip"`
}

func (a TaskAsset) GetAsset() string {
	return a.Name
}

func (a TaskAsset) GetChecksum() Checksum {
	return a.Checksum
}

func (a TaskAsset) GetUnzip() bool {
	return a.Unzip
}
