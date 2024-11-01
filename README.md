# OBS Controller

[![Build Status](https://github.com/dustinbowers/obs-controller/actions/workflows/build.yml/badge.svg)](https://github.com/dustinbowers/obs-controller/actions)  

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


# WebSocket Protocol

## Data Envelope

All messages are wrapped in this "action envelope" format

```
{
  "action": "action_name",
  "data": {
    ...
  }
}
```

## Actions

### Ping

**Action**: `ping`   
**Sender**: *streamerApp*   
**Receiver**: *extension*
```
{
  "action": "ping",
  "data": {}
}
```

### Update Video Output Settings 

**Action**: `update_video_settings`   
**Sender**: *streamerApp*  
**Receiver**: *extension*
```
{
  "action": "update_video_settings",
  "data": {
    "output_width": 123,
    "output_height": 123
  }
}
```

### Update SceneItem Positions

**Action**: `update_scene_items`   
**Sender**: *streamerApp*  
**Receiver**: *extension*
```
{
  "action": "update_scene_items",
  "data": {
    "items": [
      {
        "id": 1,
        "x": 123.5,
        "y": 123.5,
        "width": 100.5,
        "height": 100.5,
        "info": "some_meta_data",
        "z_index": 10
      }
    ]
  }
}
```

### Update SceneItem Bounds  

**Action**: `update_bounds`   
**Sender**: *streamerApp*  
**Receiver**: *extension*
```
{
  "action": "update_bounds",
  "data": {
    "bounds": {
      "84": {
        "left": 0.25,
        "top": 0,
        "right": 1,
        "bottom": 0.75
      },
      "96": {
        "left": 0.45,
        "top": 0.25,
        "right": 1,
        "bottom": 0.9
      }
    }
  }
}
```

### Move OBS SceneItems
```diff
- needs refactoring
```
**Action**: `set_scene_item_transform`   
**Sender**: *extension*  
**Receiver**: *streamerApp*
```
{
  "action": "update_scene_items",
  "data": [
    {
      "data": [
        {
          "name": 1,
          "x": 286.4164733886719,
          "y": 104.30857849121094,
          "width": "340.000000",
          "height": "340.000000",
          "info": "some data to register later",
          "zIndex": 10
        }
      ]
    },
    {
      "data": [
        {
          "name": 2,
          "x": 898.9730224609375,
          "y": 91.8488540649414,
          "width": "289.000000",
          "height": "289.000000",
          "info": "some data to register later",
          "zIndex": 10
        }
      ]
    }
  ]
}
```
