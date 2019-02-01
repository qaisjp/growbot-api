# sdp-web-server
The web server for GrowBot Web.

## Contributing & Installing

[Go installation instructions]: https://golang.org/doc/install

1. Make sure Go is set up (you should have `GOROOT` and `GOPATH` set up appropriately).

    Don't have Go? Please follow the [Go installation instructions].

    You should also add `$GOPATH/bin` to your `$PATH`. This is so that any self-built binaries (such as [`goimports`](https://godoc.org/golang.org/x/tools/cmd/goimports)) are easy to run.

1. Already have Go set up? Make sure `go version` says you are running go 1.11 or later. We use [the new Go Modules system](https://blog.golang.org/modules2019), so if necessary, follow the [Go installation instructions] to upgrade.

    Curious about Go Modules? See the Go [Modules](https://github.com/golang/go/wiki/Modules) wiki page!

1. Make sure your editor is set up correctly. All files should be correctly formatted as per `go fmt`.

    We recommend you to use any of the following extensions with the obvious editors: [vscode-go](https://code.visualstudio.com/docs/languages/go), [GoSublime](https://packagecontrol.io/packages/GoSublime) or [vim-go](https://github.com/fatih/vim-go). These all run `gofmt` whenever you save a `.go` file. (You may find using `goimports` instead of `gofmt` more convenient.)

    If you choose to use Goland, please:
    - **do not** add any IDEA-related workspace folder to Git.
    - configure `gofmt` on save—it is not a default feature. Here are [instructions](https://stackoverflow.com/questions/33774950/execute-gofmt-on-file-save-in-intellij) to configure `gofmt` on file save. (This Stack Overflow link has not been tested, so you may find other instructions online that are easier to follow.)

1. Git commits must be [clean, atomic,](https://chris.beams.io/posts/git-commit/) and most importantly, follow the [seven sacred rules](https://chris.beams.io/posts/git-commit/#seven-rules).

    If you have a big commit that needs splitting up, you can use `git add --patch` (in short, `git add -p`) to interactively stage [hunks](https://www.bignerdranch.com/blog/using-git-hunks/).

    If you are stuck, please ask for assistance!

1. In your folder of choice run `git clone github.com/qaisjp/sdp-web-server`.

    If you clone the repository inside your `$GOPATH`, make sure the root of the repository is situated at exactly `$GOPATH/src/github.com/teamxiv/growbot-api`.

    If you clone the repository anywhere else on your filesystem, the folder does not matter.

    _Why does the folder name matter at all?_ The actual _package_ name is `github.com/teamxiv/growbot-api` whereas for now the _repository_ is situated at `qaisjp/sdp-web-server`. These will both align when the repository is made public. Paths inside `$GOPATH` determine packages, elsewhere (i.e. not in `$GOPATH`) the package name is determined by the `go.mod` file.

1. Once you have the repository cloned, type `go mod download` to download dependencies. Then type `go mod verify` to verify that all is OK.

    If you run into a `command not found` error, please jump to step 1 and re-read the [Go installation instructions].

1. Now, from any directory, you can run the following command: `go install github.com/teamxiv/growbot-api/cmd/growbot-api`.
    - This will build the command-line program (from [`./cmd/growbot-api`](/cmd/growbot-api/main.go)) to your `$GOPATH/bin` directory (in step 1 you should have added this path to your `$PATH`).
    - You can now simply type `growbot-api` from any directory to start the API.

## Running

Command line documentation is available by supplying the `--help` argument. As of 737fa69e9799f8886103180ed318395b8a863c96 the following text is printed:

```
➜ growbot-api --help
Usage of growbot-api:
  -bindaddress
    	Change value of BindAddress. (default 0.0.0.0:8080)
  -loglevel
    	Change value of LogLevel. (default debug)

Generated environment variables:
   CONFIG_BINDADDRESS
   CONFIG_LOGLEVEL

flag: help requested
```

Refer to [./internal/config](/internal/config/config.go) to see what each of these options mean.

Please note that `0.0.0.0:8080` means "bind to all addresses, port 8080". You should try to connect to your loopback address instead of `0.0.0.0:8080`. This is usually `127.0.0.1` or `localhost`.

Instead of using environment variables or command-line arguments, you can use a YAML, JSON or TOML file.
- Just use the argument names as a data key. See `config.example.yml` as an example.
- To use a config file, e.g. `config.yml`, set the `config` environment variable, like so: `config=config.yml growbot-api`.

### Test websockets

You can use [wsc](https://github.com/danielstjules/wsc). Just do `yarn global add wsc` and then `wsc -er "ws://localhost:8080/stream/<uuid>` should work!

## Nomenclature

- `uuid`s are provided by [`github.com/google/uuid`](https://godoc.org/github.com/google/uuid).

    UUIDs will generally be used as the serial number of the robot. User accounts will most likely use an `int64` as their primary key.

## Endpoints

The API is not stable as this project is very much in rapid development.

### Stage 1

This is a very primitive version of the API with minimal care for security. The aim here is to get a proof-of-concept version of the API functioning.

#### POST `/move`

Requests a specific robot to move in a certain direction.

Authenticated: _true_

Query parameters: _none_

Headers expected:

- `Content-Type: application/json`

Payload:

```json
{
    "id": "uuid", // robot uuid
    "direction": "(forward|backwards|left|right)"
}
```

Responses:

- 400 Bad Request: if input does not match above schema
- 404 Not Found: if the robot is not connected
- 200 OK: if the request was sent to the robot (receival is not guaranteed)

#### GET `/stream/<uuid>` (websocket)

This endpoint is for robots to receive live updates to the API. The website will use a separate websocket endpoint that is not limited to one robot.

Authenticated: _false_

Headers expected (usually supplied by your websocket library):

- `Connection: Upgrade`
- `Upgrade: websocket`

Parameters:

- `uuid` is a uuid of the current robot.

Clients connected to this websocket server will receive updates in the following format:

```json
{
    "type": string,
    "data": <any>
}
```

##### type: `move`, data: `string`

This data will just be a string contain the cardinal direction `north` or `south`.

### Stage 2

All, or a subset of, the following tasks:

1. Pin down security practices
1. Add database schema
1. Add user system
1. Allow users to register robots to their account
1. Streaming data back to the central server
1. ???

## Software License

This software is governed by the license defined in [/LICENSE](/LICENSE) at the root of this repository.
