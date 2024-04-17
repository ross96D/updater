port:            11111
user_secret_key: ""
user_jwt_expiry: "2m"
apps: [
	{
		owner:                 "ross96D"
		repo:                  "updater"
		github_webhook_secret: "-"
		github_auth_token:     "-"
		asset_name:            "-"
		task_sched_path:       "-"
		system_path:           "-"
		checksum: {
			direct_asset_name: "-"
		}
		additional_assets: [
			{
				name:        "asset1"
				system_path: "path1"
				checksum: {
					direct_asset_name: "-"
				}
			},
			{
				name:        "asset1"
				system_path: "path1"
				checksum: {
					direct_asset_name: "-"
				}
			},
		]
	},
	{
		owner:                 "ross96D"
		repo:                  "updater2"
		github_webhook_secret: "-"
		github_auth_token:     "-"
		asset_name:            "--"
		task_sched_path:       "-"
		system_path:           "-"
		checksum: {
			direct_asset_name: "-"
		}
	},
]
