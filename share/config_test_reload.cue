port:            1234
user_secret_key: "some_key"
user_jwt_expiry: "2h"
apps: [
	{
		owner:               "ross96D"
		repo:                "updater"
		github_signature256: "sign"
		github_auth_token:   "auth"
		asset_name:          "some asset name"
		task_sched_path:     "/is/a/path"
		app_path:            "/is/a/path"
	},
]