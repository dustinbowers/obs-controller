package types

// Bound represents the bounds of a window.
type Bound struct {
	Left   float64 `json:"left"`
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
}

type Config struct {
	ObsHost        string `toml:"obs_host"`
	ObsPort        string `toml:"obs_port"`
	ObsPassword    string `toml:"obs_password"`
	TwitchUsername string `toml:"twitch_username"`
}

// WindowConfig represents the structure of windowConfig.json.
type WindowConfig struct {
	Bounds map[string]Bound `json:"bounds"`
}

// Info represents each entry in the infoWindow map with title and description fields.
type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// InfoWindowData represents the structure of infoWindowData.json.
type InfoWindowData struct {
	InfoWindow map[string]Info `json:"infoWindow"`
}
