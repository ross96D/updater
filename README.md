# Upgrade your vps services from github workflows

Updater is a tool that allows to manage the update of services in a vps. Currently there are 2
ways to update a service. Using an endpoint (design to be called from a build workflow) or using
a github repo release.

Also there is a tui app to interact with several updater services from you local computer.

## Updater flags

- `-c`, `--config` set the path to the configuration file. If the path is not an absolute path
  the the current working directory will be used (default "config.cue")
- `--cert` the tls certificate path to be used in https. If not present the server will use http
- `--key` the tls key to be used in https. If not present the server will use http

## Setting a configuration file

Below is the schema file that updater uses to validate the configuration.

- `!` at the end of variable name means required: `varname!`
- `?` at the end of variable name means optional: `varname?`
- `[...#Type]` array of elements with type `#Type`
- `time.Duration()` this is from golang time.ParseDuration documentation: A duration string is a
  possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix,
  such as "300ms", "-1.5h" or "2h45m". Valid time units are
  `ns` nanoseconds, `us` (or `µs`) microseconds, `ms` miliseconds, `s` seconds, `m` minute, `h` hour

```cue
import "time"

port!:                  uint16              // port where the updater will listen
user_secret_key!:       string              // the key used to encode the json web token. Used to authenticate the user endpoints
user_jwt_expiry!:       time.Duration()     // the time to expire the user json web token
users:                  [...#User]          // list of user allowed to interface with the updater
apps:                   [...#Application]   // list of the apps that the updater will update
base_path?:             string              // path where the temporal files used by the app will place

// user credentials, represent a user that will be allowed to interact with the updater
#User: {
 name!:     string
 password!: string
}

#Application: {
 name?:         string              // this name is only used for human readability
 auth_token?:   string              // token that identify an application and authorize a user to update
 assets!:       [...#Asset]         // Assets to update

 // service path used for systemd/task-scheduler.
 // if set the service will be stopped at the beggining of the asset update and restarted at the end
 service?:      string

 cmd?: #Command                     // command to run after the application update all his assets

 github_release?: #GithubRelease    // github repository where to find the latest release for manual application update
}

#GithubRelease: {
 token?: string // set if the repo is not a public one
 repo!:  string // repository name <github.com/$owner/$repo>
 owner!: string // repository owner name <github.com/$owner/$repo>
}

#Asset: {
 name!:        string       // the name of the form field
 system_path!: string       // path where the asset should be included
 keep_old: bool | *false    // (default false) if true keeps the previous version with at .old at the end

 // service path used for systemd/task-scheduler.
 // if set the service will be stopped at the beggining of the asset update and restarted at the end
 service?:     string

 unzip: bool | *false       // (default false) if this true, the asset will be decompressed
 cmd?:  #Command            // command to run after the asset is copy
}

#Command: {
 command!:  string      // the command name or absolute path to binary
 args?:     [...string] // arguments to be passed to the command invocation

 // working directory where the command should be executed.
 // if no value is provided the working directory of updater is used
 path?:     string
}

```

### Configuration example

```cue
port:            7432
user_secret_key: "super secret key"
user_jwt_expiry: "2h"
users: [
    {
        name:     "my-name"
        password: "my-password"
    },
]

apps: [
    {
        name:       "my-app"
        auth_token: "secure-token-my-app"
        assets: [
            {
                name:        "app"
                service:     "my-app.service"
                system_path: "/usr/bin/myapp"
                unzip:       true
            },
            {
                name:        "assets"
                system_path: "/usr/share/my-app/assets.tar.gz"
                unzip:       true
                cmd: {
                    command: "ls"
                    args:    ["--all"]
                    path:    "/usr/share/my-app/"
                }
            },
        ]
    },
]
```

## Client

Rigth now there is a desktop client in development, see [here](https://github.com/ross96d/updater_client)
