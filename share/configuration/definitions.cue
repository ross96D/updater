import "time"

port!:            uint16 // port where the updater will listen
user_secret_key!: string // the key used to encode the json web token. Used to authenticate the user endpoints
user_jwt_expiry!: time.Duration() // the time to expire the user json web token
users:            [...#User] // list of user allowed to interface with the updater
apps: [...#Application] // list of the apps that the updater will update
base_path?: string // path where the temporal files used by the app will place

// user credentials, represent a user that will be allowed to interact with the updater
#User: {
    name!: string
    password!: string
}

#Application: {
    auth_token?: string

    assets?: [...#Asset]

    // use this to set a command to be run after succesfully update 
    post_action?: #Command
}

#Asset: {
    // the name of the form field
    name!: string
    service?: string
    system_path!: string

    // if this is set to true, the asset will be decompressed
    unzip: bool | *false
}

#Command: {
    command!: string
    args?: [...string]
}
