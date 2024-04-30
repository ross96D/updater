import "time"

port!:            uint16 // port where the updater will listen
user_secret_key!: string // the key used to encode the json web token. Used to authenticate the user endpoints
user_jwt_expiry!: time.Duration() // the time to expire the user json web token
users:            [...#User] // list of user allowed to interface with the updater
apps: [...#Application] // list of the apps that the updater will update
base_path?: string // path where the temporal files used by the app will place

// The application represent a repo to recieve updates.
// For now also represent a asset that have a task path
#Application: {
    owner!: string
    repo!: string
    host: string | *"github.com" // if not set will have the value github.com
    github_webhook_secret!: string
    github_auth_token?: string
    
    asset_name!: string
    task_sched_path!: string
    system_path!: string

    // if not set will have the value NoChecksum
    checksum: #DirectChecksum | #AggregateChecksum | #CustomChecksum | *#NoChecksum

    additional_assets?: [...#AdditionalAsset]

    // use this to set a command to be run after succesfully update 
    post_action?: #Command

    // use cache allow to check against the checksum upstream and if it match the file will not downloaded
    use_cache: bool | *true 
}

// user credentials, represent a user that will be allowed to interact with the updater
#User: {
    name!: string
    password!: string
}

// The additional asset is used to add files that are on the repo but are not asociated with a task
// on the task schedule
#AdditionalAsset: {
    name: string
    system_path: string
    checksum: #DirectChecksum | #AggregateChecksum | #CustomChecksum | *#NoChecksum
}

// the direct checksum search for the direct_asset_name on the repo release, and if is found then
// use the content of the file as the checksum
#DirectChecksum: {
    type: null | *"DirectChecksum"
    direct_asset_name!: string
}


// the aggregate checksum search for the aggregate_asset_name on the release, and if is found then
// search for the (AggregateChecksum.Key ?? Application.AssetName) and use the hash for that
//
// The expected format is:
//
// <hexadecimal encoded hash><blank space><key1>
//
// <hexadecimal encoded hash><blank space><key2>
#AggregateChecksum: {
    type: null | *"AggregateChecksum" 
    aggregate_asset_name!: string
    key?: string
}

// The custom checksum let the user make a custom script for the checksum implementation.
// The script recieves the github token as an enviroment variable named "__UPDATER_GTIHUB_TOKEN"
//
// The script should output the hash value on the stdout as a hexadecimal encoded string.
#CustomChecksum: {
    type: null | *"CustomChecksum" 
    command!: string
    args?: [...string]
}

// No checksum will no perform any checksum operation
#NoChecksum: {}

#Command: {
    command!: string
    args?: [...string]
}
