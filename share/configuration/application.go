package configuration

type IRepo interface {
	GetRepo() (host, owner, repo string)
}

type Application struct {
	AuthToken string `json:"auth_token"`

	Assets []Asset `json:"assets"`

	PostAction *Command `json:"post_action"`
}

type Command struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}
