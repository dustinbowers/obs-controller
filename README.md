# OBS Controller

![Build action](https://github.com/dustinbowers/obs-controller/actions/workflows/go.yml/badge.svg)

A Go port of the MatisseTec streamerApp

## Usage


### Setup:

Clone the repo and then copy `example_config.toml` to `config.toml`

Edit `config.toml` to set your own `twitch_username` and `obs_password` fields

```toml
twitch_username = "YOUR_TWITCH_USERNAME"
obs_password = "YOUR_OBS_PASSWORD"

obs_host = "localhost"
obs_port = "4455"
```

### Build the binary:
```
go build
```

### Or Run directly:

```
go run main.go
```

