package controller

import (
	"fmt"
	"io"
	"net/http"
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
