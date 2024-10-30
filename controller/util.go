package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"obs-controller/controller/types"
	"os"
)

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

// SaveInfoWindowData marshals InfoWindowData struct and saves it to a JSON file.
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
