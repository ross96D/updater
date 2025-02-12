package configuration

type IRepo interface {
	GetRepo() (host, owner, repo string)
}

type Application struct {
	Name string `json:"name"`

	AuthToken string `json:"auth_token"`

	Service string `json:"service"`

	ServiceType string `json:"service_type"`

	Assets []Asset `json:"assets"`

	AsstesOrder []AssetOrder

	AssetsDependency map[string][]string `json:"assets_dependency"`

	Command *Command `json:"cmd"`

	GithubRelease *GithubRelease `json:"github_release"`
}

type GithubRelease struct {
	Token string `json:"token"`
	Repo  string `json:"repo"`
	Owner string `json:"owner"`
}
