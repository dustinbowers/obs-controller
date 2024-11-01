package main

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"io"
	"log"
	"net/http"
	"obs-controller/controller"
	"os"
)

type Config struct {
	ObsHost        string `toml:"obs_host"`
	ObsPort        string `toml:"obs_port"`
	ObsPassword    string `toml:"obs_password"`
	TwitchUsername string `toml:"twitch_username"`
}

// LoadConfig reads and parses the TOML config file into a Config struct.
func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := toml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

func GetTwitchUserID(twitchUsername string) (string, error) {
	requestUrl := fmt.Sprintf("https://decapi.me/twitch/id/%s", twitchUsername)
	response, err := http.Get(requestUrl)
	if err != nil {
		return "", fmt.Errorf("GET request failed: %v", err)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}
	return string(body), nil
}

func main() {
	log.Printf("Welcome Streamer!\n")

	// Load configs from file
	config, err := LoadConfig("config.toml")
	if err != nil {
		log.Fatalf("Failed to load config.toml: %s", err)
	}

	// Fetch twitch user ID from username
	twitchUserId, err := GetTwitchUserID(config.TwitchUsername)
	if err != nil {
		log.Fatalf("Failed to get twitch user id: %s\n", err)
	}

	// Create the ObsController that holds the OBS and Web proxy websocket connections
	ctl, err := controller.NewController(
		fmt.Sprintf("%s:%s", config.ObsHost, config.ObsPort),
		config.ObsPassword,
		twitchUserId)
	if err != nil {
		log.Fatalf("Failed to create OBS controller: %s", err)
	}
	defer ctl.Cleanup()

	// Start the main listen-parse-update event loop
	err = ctl.Run()
	if err != nil {
		log.Fatalf("OBS Controller Error: %s\n", err)
	}
}
