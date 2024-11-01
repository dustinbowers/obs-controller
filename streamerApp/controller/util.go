package controller

import (
	"encoding/json"
	"fmt"
	"github.com/pelletier/go-toml"
	"io"
	"net/http"
	"obs-controller/controller/types"
	"os"
)

// GetRoomKey fetches the key for the websocket room from the web proxy
func GetRoomKey(postUrl string) (string, error) {
	// Fetch key for new room
	response, err := http.Post(postUrl, "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("POST request failed: %v", err)
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}
	return string(body), nil
}

// ReadWindowConfig reads and parses the JSON configuration file into a WindowConfig struct.
func ReadWindowConfig(filename string) (*types.WindowConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config types.WindowConfig
	if err := json.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// ReadInfoWindowData loads InfoWindowData from a file
func ReadInfoWindowData(filename string) (*types.InfoWindowData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var data types.InfoWindowData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// SaveInfoWindowData marshals InfoWindowData struct and writes it to disk
func SaveInfoWindowData(filename string, data *types.InfoWindowData) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filename, bytes, 0644); err != nil {
		return err
	}

	return nil
}

// LoadConfig reads and parses the TOML config file into a UserConfig struct.
func LoadConfig(filename string) (*types.Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config types.Config
	decoder := toml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SaveConfig writes the given UserConfig struct to a TOML config file.
func SaveConfig(filename string, config *types.Config) error {
	// Create or truncate the config file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Encode the UserConfig struct and write it to the file
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return err
	}

	return nil
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
