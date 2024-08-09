port:            11111
user_secret_key: ""
user_jwt_expiry: "2m"
apps: [
	{
		auth_token: "-"

		assets: [
			{
				name:            "-"
				task_sched_path: "-"
				system_path:     "-"
			},
			{
				name:        "asset1"
				system_path: "path1"
			},
			{
				name:        "asset1"
				system_path: "path1"
			},
		]
	},
	{
		auth_token: "-"

		assets: [
			{
				name:            "--"
				task_sched_path: "-"
				system_path:     "-"
			},
		]
	},
]
