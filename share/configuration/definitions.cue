import "time"

#User: {
    name!: string
    password!: string
}

#DirectChecksum: {
    type: null | *"DirectChecksum"
    direct_asset_name!: string
}

#AggregateChecksum: {
    type: null | *"AggregateChecksum" 
    aggregate_asset_name!: string
    key?: string
}

#CustomChecksum: {
    type: null | *"CustomChecksum" 
    command!: string
    args?: [...string]
}

#NoChecksum: {}

#Application: {
    owner!: string
    repo!: string
    host: string | *"github.com" 
    github_signature256!: string
    github_auth_token?: string
    
    asset_name!: string
    task_sched_path!: string
    app_path!: string
    
    checksum: #DirectChecksum | #AggregateChecksum | #CustomChecksum | *#NoChecksum

    use_cache: bool | *true 
}


port!:            uint16
user_secret_key!: string
user_jwt_expiry!: time.Duration()
users:            [...#User]
apps: [...#Application]
base_path?: string