port:            1234
user_secret_key: "some_key"
user_jwt_expiry: "2h"
apps: [
	{
		auth_token: "auth"

		assets: [
			{
				name:        "some asset name"
				service:     "/is/a/path"
				system_path: "/is/a/path"
			},
		]

		cmd: {
			command: "python"
			args: ["-f", "-s"]
		}
	},
]
