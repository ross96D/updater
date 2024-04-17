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

#AdditionalAsset: {
    name: string
    system_path: string
    checksum: #DirectChecksum | #AggregateChecksum | #CustomChecksum | *#NoChecksum
}

#NoChecksum: {}

#Application: {
    owner!: string
    repo!: string
    host: string | *"github.com" 
    github_webhook_secret!: string
    github_auth_token?: string
    
    asset_name!: string
    task_sched_path!: string
    system_path!: string
    
    checksum: #DirectChecksum | #AggregateChecksum | #CustomChecksum | *#NoChecksum

    additional_assets?: [...#AdditionalAsset]

    use_cache: bool | *true 
}


port!:            uint16
user_secret_key!: string
user_jwt_expiry!: time.Duration()
users:            [...#User]
apps: [...#Application]
base_path?: string