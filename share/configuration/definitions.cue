import "time"

port!:            uint16          // port where the updater will listen
user_secret_key!: string          // the key used to encode the json web token. Used to authenticate the user endpoints
user_jwt_expiry!: time.Duration() // the time to expire the user json web token
users: [...#User] // list of user allowed to interface with the updater
apps: [...#Application] // list of the apps that the updater will update
base_path?:             string // path where the temporal files used by the app will place

// user credentials, represent a user that will be allowed to interact with the updater
#User: {
	name!:     string
	password!: string
}

#Application: {
	name?:         string
	auth_token?:   string
	service?:      string
	service_type?: "nssm" | "taskservice"
	assets!: [...#Asset]

	// Declares an assets dependency.
	// write it like:
	//     assets_dependency: {
	//         "asset1": ["asset2", "asset3"]
	//         "asset2": ["asset3"]
	//     }
	// make sure not to write a cyclic dependency
	assets_dependency?: [string]: [...string]

	// use this to set a command to be run after succesfully update
	cmd?: #Command

	github_release?: #GithubRelease
}

#GithubRelease: {
	token?: string
	repo!:  string
	owner!: string
}

#Asset: {
	// the name of the form field
	name!:         string
	service?:      string
	service_type?: "nssm" | "taskservice"
	system_path!:  string

	// if keeps the previous version with at .old at the end
	keep_old: bool | *false

	// if this is set to true, the asset will be decompressed
	unzip: bool | *false
	cmd?:  #Command
}

#Command: {
	command!: string
	args?: [...string]
	path?: string

	// When the command timeout the execution will continue assuming command success
	// and the command output should be given to the in another form.
	//
	// This means that the on timeout the command must not be killed.
	timeout: time.Duration() | *"5m"
}
