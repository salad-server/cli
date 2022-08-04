# Salad-cli
![go-version](https://img.shields.io/github/go-mod/go-version/salad-server/cli) ![report-card](https://goreportcard.com/badge/github.com/salad-server/cli) ![last-commit](https://img.shields.io/github/last-commit/salad-server/cli)

Salad CLI is a simple command line interface for managing the salad stack. **This is NOT a replacement for an admin panel.** The purpose of this CLI is to do small time maintenance jobs such as updating qualified maps to ranked, or backing up the server.

## Requirements
- **Go 1.18** is required.
- **tmux** is required for process management.
- **make and git** are also required, but on most systems these are installed by default.
- **[upx](https://upx.github.io/)** is optional, but recommended as it will compress the proxy.


## Install
Clone and build the source:
```sh
$ git clone https://github.com/salad-server/cli.git
$ cd cli
$ make build
$ make build-prod # OPTIONAL: upx alternative
```

Configure `config.yaml`:
```sh
$ cp config.example.yaml config.yaml
$ nano config.yaml
```

**OPTIONAL:** Link to `/bin` directory:
```sh
$ sudo ln -r -s cli /bin/salad
```

This will make the CLI executable from anywhere. Just note that backup archives are still stored in the `/backup` directory.

## Usage
Update beatmap database:
```sh
# Usage
$ salad help update

# Update all qualified maps
$ salad update -s qualified

# Update beatmapset <id>
$ salad update -b 30682
```

Backup:
```sh
# Usage
$ salad help backup

# Backup (use args to ignore)
$ salad backup
$ salad backup --sql --replays # Exclude SQL and replays
```

Personal best:
```sh
# Usage
$ salad help pb

# Mark score <id> as personal best
$ salad pb 1
```

Process management:
```sh
# Usage
$ salad help process

# Create tmux session with process list from config.yaml
$ salad --start
$ salad --start -a # to not attach once finished

# Gracefully shutdown all processes in tmux session
$ salad --stop

# Restart (trigger graceful stop if session is active, start new session)
$ salad --restart
$ salad --start -a # to not attach once finished
```

Makefile:
```sh
# Run without building
# Pass arguments through args
$ make run args="process --stop"

# Build
$ make build

# Build (with upx)
$ make build-prod
```

## License

  [MIT](LICENSE)
