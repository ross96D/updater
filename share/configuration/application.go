package configuration

import "strings"

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

type Command struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Path    string   `json:"path"`
}

func (c Command) String() string {
	builder := strings.Builder{}
	if c.Path != "" {
		builder.WriteString(c.Path + ": ")
	}
	builder.WriteString(c.Command + " ")
	builder.WriteString(strings.Join(c.Args, " "))
	return builder.String()
}
