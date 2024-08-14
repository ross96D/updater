package configuration

type IRepo interface {
	GetRepo() (host, owner, repo string)
}

type Application struct {
	AuthToken string `json:"auth_token"`

	Assets []Asset `json:"assets"`

	PostAction *Command `json:"post_action"`

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
}
