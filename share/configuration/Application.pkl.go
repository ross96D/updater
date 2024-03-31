// Code generated from Pkl module `updater.share.Configuration`. DO NOT EDIT.
package configuration

type Application struct {
	Owner string `pkl:"owner"`

	Repo string `pkl:"repo"`

	Host string `pkl:"host"`

	GitubSignature256 string `pkl:"gitub_signature256"`

	GithubAuthToken string `pkl:"github_auth_token"`

	AssetName string `pkl:"asset_name"`

	TaskSchedPath string `pkl:"task_sched_path"`

	AppPath string `pkl:"app_path"`

	Checksum any `pkl:"checksum"`

	UseCache bool `pkl:"use_cache"`
}
