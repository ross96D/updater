port:            11111
user_secret_key: ""
user_jwt_expiry: "2m"
apps: [
	{
		auth_token: "-"
		service:    "nothing"

		assets: [
			{
				name:        "asset0"
				service:     "-"
				system_path: "-"
			},
			{
				name:        "asset1"
				system_path: "path1"
			},
			{
				name:        "asset2"
				system_path: "path1"
			},
		]
	},
	{
		auth_token: "-"

		assets: [
			{
				name:        "--"
				service:     "-"
				system_path: "-"
				cmd: {
					command: "cmd"
					args: ["arg1", "arg2"]
					path: "some/path"
				}
			},
		]
	},
]
