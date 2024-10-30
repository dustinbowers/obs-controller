package controller

import (
	"fmt"
	"github.com/andreykaipov/goobs"
	"log"
)

type ObsController struct {
	ObsClient *goobs.Client
	WebClient *WebClient
}

func NewController(obsHost string, obsPassword string, webUserId string) (*ObsController, error) {
	// Connect to OBS
	log.Printf("Connecting to OBS...\n")
	obsClient, err := goobs.New(obsHost, goobs.WithPassword(obsPassword))
	if err != nil {
		return nil, err
	}
	log.Println("Done")

	// Fetch room key for user
	log.Printf("Fetching room key for %s\n", webUserId)
	newRoomUrl := fmt.Sprintf("https://websocket.matissetec.dev/lobby/new?user=%s", webUserId)
	roomKey, err := GetRoomKey(newRoomUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to get room key for user %s: %v", webUserId, err)
	}
	log.Printf("Room key: %s\n", roomKey)

	// Connect to Websocket Proxy
	log.Printf("Connecting to websocket proxy...\n")
	wsAddr := fmt.Sprintf("wss://websocket.matissetec.dev/lobby/connect/streamer?user=%s&key=%s", webUserId, roomKey)
	webClient, err := NewWebClient(wsAddr)
	if err != nil {
		return nil, fmt.Errorf("connection to websocket proxy failed: %v", err)
	}
	log.Printf("Done\n")

	newClient := ObsController{
		ObsClient: obsClient,
		WebClient: webClient,
	}
	return &newClient, nil
}

func (ctl *ObsController) Cleanup() error {
	return ctl.ObsClient.Disconnect()
}

func (ctl *ObsController) PrintObsVersion() error {
	version, err := ctl.ObsClient.General.GetVersion()
	if err != nil {
		return err
	}

	fmt.Printf("OBS Studio version: %s\n", version.ObsVersion)
	fmt.Printf("Server protocol version: %s\n", version.ObsWebSocketVersion)
	fmt.Printf("Client protocol version: %s\n", goobs.ProtocolVersion)
	fmt.Printf("Client library version: %s\n", goobs.LibraryVersion)

	return nil
}

func (ctl *ObsController) Run() error {
	log.Printf("OBS Controller running...")
	for {
		select {
		case close := <-ctl.WebClient.Close:
			log.Printf("Websocket proxy closed: %v", close)
			return nil
		case message := <-ctl.WebClient.Message:
			log.Printf("Websocket proxy message: %v", message)
		}
	}
}
