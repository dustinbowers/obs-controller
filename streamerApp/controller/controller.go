package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/config"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"github.com/andreykaipov/goobs/api/typedefs"
	"log"
	"obs-controller/controller/types"
	"slices"
	"strings"
	"time"
)

// ObsController holds the OBS and Web proxy connections
type ObsController struct {
	ObsClient *goobs.Client
	WebClient *WebClient

	UserConfig   *types.Config
	WindowConfig *types.WindowConfig

	stopNetworkCtx  *context.Context
	stopNetworkFunc *context.CancelFunc

	connectionStatus string
}

func NewController() (*ObsController, error) {
	userConfig, err := LoadConfig("config.toml")
	if err != nil {
		return nil, fmt.Errorf("failed to load config.toml: %v", err)
	}

	// Load WindowConfig
	windowConfig, err := ReadWindowConfig("windowConfig.json")
	if err != nil {
		return nil, fmt.Errorf("read window config err: %v", err)
	}

	newClient := ObsController{
		ObsClient:        nil,
		WebClient:        nil,
		UserConfig:       userConfig,
		WindowConfig:     windowConfig,
		connectionStatus: "Disconnected",
	}
	return &newClient, nil
}

func (ctl *ObsController) Start() error {
	ctl.connectionStatus = "Connecting..."
	if ctl.stopNetworkFunc != nil {
		(*ctl.stopNetworkFunc)()
	}
	stopNetworkCtx, stopNetworkFunc := context.WithCancel(context.Background())
	ctl.stopNetworkCtx = &stopNetworkCtx
	ctl.stopNetworkFunc = &stopNetworkFunc

	// Connect to OBS
	obsHost := fmt.Sprintf("%s:%s", ctl.UserConfig.ObsHost, ctl.UserConfig.ObsPort)
	log.Printf("Connecting to OBS...")
	obsClient, err := goobs.New(obsHost, goobs.WithPassword(ctl.UserConfig.ObsPassword))
	if err != nil {
		return err
	}
	log.Println("Done")
	if err := PrintObsVersion(obsClient); err != nil {
		return fmt.Errorf("error printing obs version: %v", err)
	}
	ctl.ObsClient = obsClient

	// Fetch twitch user ID from username
	twitchUserId, err := GetTwitchUserID(ctl.UserConfig.TwitchUsername)
	if err != nil {
		log.Fatalf("Failed to get twitch user id: %s\n", err)
	}

	// Fetch room key for user
	log.Printf("Fetching room key for %s\n", ctl.UserConfig.TwitchUsername)
	newRoomUrl := fmt.Sprintf("https://websocket.matissetec.dev/lobby/new?user=%s", twitchUserId)
	roomKey, err := GetRoomKey(newRoomUrl)
	if err != nil {
		return fmt.Errorf("failed to get room key for user %s: %v", twitchUserId, err)
	}
	log.Printf("\tRoom key: %s\n", strings.Repeat("*", len(roomKey)))

	// Connect to Websocket Proxy
	log.Printf("Connecting to proxy...")
	wsAddr := fmt.Sprintf("wss://websocket.matissetec.dev/lobby/connect/streamer?user=%s&key=%s", twitchUserId, roomKey)
	webClient, err := NewWebClient(wsAddr)
	if err != nil {
		return fmt.Errorf("connection to proxy failed: %v", err)
	}
	ctl.WebClient = webClient
	log.Printf("Done\n")

	go func() {
		err := ctl.run(*ctl.stopNetworkCtx)
		ctl.connectionStatus = "Disconnected"
		if err != nil {

			log.Println(err)
		}
	}()

	return nil
}

func (ctl *ObsController) Cleanup() error {
	if ctl.stopNetworkFunc != nil {
		(*ctl.stopNetworkFunc)()
	}
	// This looks convoluted, but...
	// the idea is "don't return early without trying to clean up both connections
	var err1, err2 error
	if ctl.ObsClient != nil {
		err1 = ctl.ObsClient.Disconnect()
	}
	if ctl.WebClient != nil {
		err2 = ctl.WebClient.Disconnect()
	}
	if err1 != nil {
		return err1
	} else if err2 != nil {
		return err2
	}
	return nil
}

func PrintObsVersion(client *goobs.Client) error {
	version, err := client.General.GetVersion()
	if err != nil {
		return err
	}
	log.Printf("OBS Studio version: %s\n", version.ObsVersion)
	log.Printf("Server protocol version: %s\n", version.ObsWebSocketVersion)
	log.Printf("Client protocol version: %s\n", goobs.ProtocolVersion)
	log.Printf("Client library version: %s\n", goobs.LibraryVersion)
	return nil
}

func (ctl *ObsController) run(ctx context.Context) error {
	log.Printf("Starting Webclient read pump...")

	// All this function does is listen for incoming web proxy messages, and sends them
	// into a buffered channel that we can process later in this function
	// NOTE: this is a go-routine, so this function continues running in its own thread
	go ctl.WebClient.StartReadPump(*ctl.stopNetworkCtx)
	log.Printf("Running")

	videoOutputSettings, err := ctl.GetVideoOutputSettings()
	if err != nil {
		return err
	}

	// Send a ping to say hello
	if err := ctl.SendPing(); err != nil {
		return err
	}

	pingTicker := time.NewTicker(30 * time.Second)

	// Now start handling any update events
	log.Printf("OBS Controller running...")
	for {
		select {
		case <-pingTicker.C:
			if err := ctl.SendPing(); err != nil {
				log.Printf("error sending ping: %v", err)
			}
		case <-ctx.Done():
			return nil
		case msg := <-ctl.WebClient.Close:
			log.Printf("Websocket proxy closed: %v", msg)
			return nil
		case message := <-ctl.WebClient.Message:

			var action types.ActionEnvelope
			err := json.Unmarshal([]byte(message), &action)
			if err != nil {
				// TODO: We should probably ignore unknown messages... but I'm leaving this in for now
				log.Printf("BAD MESSAGE FORMAT: %s", message)
				break
			}
			log.Printf("INBOUND WebClient message: %v", message)

			switch action.Action {
			case "welcome":
				// Send video output settings
				if err := ctl.SendObsSizeConfig(videoOutputSettings); err != nil {
					return err
				}

				// Send window config
				if err := ctl.SendWindowConfig(); err != nil {
					return err
				}

				// Send info window data config
				if err := ctl.SendInfoWindowConfig(); err != nil {
					return err
				}

				// Send the new scene item transforms back to the web proxy
				err := ctl.SendSceneItemsToServer()
				if err != nil {
					return err
				}
			case "set_scene_item_transform":
				sceneItemTransformCommand, err := ctl.ParseSceneItemTransform(action.Data)
				if err != nil {
					log.Printf("error parsing scene item transform command: %v", err)
					break
				}
				log.Printf(" sceneItemTransformCommand: %#+v", sceneItemTransformCommand)

				// Get updated video settings
				videoOutputSettings, err := ctl.GetVideoOutputSettings()
				if err != nil {
					return err
				}

				// Get Scene Item Transform details
				currentSceneItemTransform, err := ctl.GetSceneItemTransformByID("Scene", sceneItemTransformCommand.ItemID)
				if err != nil {
					return err
				}

				currentSceneItemTransform.PositionX = sceneItemTransformCommand.X * videoOutputSettings.BaseWidth
				currentSceneItemTransform.PositionY = sceneItemTransformCommand.Y * videoOutputSettings.BaseHeight
				currentSceneItemTransform.BoundsWidth = 1.0  // TODO: Not sure why these are necessary
				currentSceneItemTransform.BoundsHeight = 1.0 // 	  but they are...

				// Send the newly received Scene item transform to OBS
				err = ctl.TransformSceneItemByID(
					"Scene",
					sceneItemTransformCommand.ItemID,
					currentSceneItemTransform)
				if err != nil {
					return err
				}

				// After the transforms are updated, we need to update the web proxy with the new coordinates
				err = ctl.SendSceneItemsToServer()
				if err != nil {
					return err
				}
			}
		}
	}
}

//
// Receiving IN from OBS
/////////////////////////

func (ctl *ObsController) GetVideoOutputSettings() (*config.GetVideoSettingsResponse, error) {
	settings, err := ctl.ObsClient.Config.GetVideoSettings()
	if err != nil {
		return nil, err
	}
	return settings, nil
}

func (ctl *ObsController) GetSelectedSceneItems() ([]int, []types.WindowDetails, error) {
	// Get SceneItems
	sceneName := "Scene"
	params := sceneitems.NewGetSceneItemListParams().WithSceneName(sceneName)
	sceneItemList, err := ctl.ObsClient.SceneItems.GetSceneItemList(params)
	if err != nil {
		return []int{}, []types.WindowDetails{}, err
	}

	// Filter to specified SceneItems
	sourceTargetNames := []string{"gitEasy", "gif", "guest1"}
	selectedIds := make([]int, 0)
	selectedItems := make([]*typedefs.SceneItem, 0)
	for _, item := range sceneItemList.SceneItems {
		if slices.Contains(sourceTargetNames, item.SourceName) {
			selectedIds = append(selectedIds, item.SceneItemID)
			selectedItems = append(selectedItems, item)
		}
	}

	// Convert the OBS response into a more manageable state
	sceneItems := make([]types.WindowDetails, 0)
	for _, item := range selectedItems {
		windowDetail := types.WindowDetails{
			SceneItemId: item.SceneItemID,
			SourceName:  item.SourceName,
			Width:       item.SceneItemTransform.Width,
			Height:      item.SceneItemTransform.Height,
			XLocation:   item.SceneItemTransform.PositionX,
			YLocation:   item.SceneItemTransform.PositionY,
		}
		sceneItems = append(sceneItems, windowDetail)
	}

	return selectedIds, sceneItems, nil
}

func (ctl *ObsController) GetSceneItemTransformByID(sceneName string, sceneItemID int) (*typedefs.SceneItemTransform, error) {
	params := sceneitems.NewGetSceneItemTransformParams().
		WithSceneName(sceneName).
		WithSceneItemId(sceneItemID)
	response, err := ctl.ObsClient.SceneItems.GetSceneItemTransform(params)
	if err != nil {
		return nil, err
	}
	return response.SceneItemTransform, nil
}

// Sending OUT to OBS
// /////////////////////

func (ctl *ObsController) TransformSceneItemByID(sceneName string, sceneItemID int, newSceneItemTransform *typedefs.SceneItemTransform) error {
	params := sceneitems.NewSetSceneItemTransformParams().
		WithSceneName(sceneName).
		WithSceneItemId(sceneItemID).
		WithSceneItemTransform(newSceneItemTransform)
	_, err := ctl.ObsClient.SceneItems.SetSceneItemTransform(params)
	if err != nil {
		return err
	}
	return nil
}

//
// Receiving IN from the webserver
// //////////////////////////////////

func (ctl *ObsController) ParseSceneItemTransform(jsonStr []byte) (*types.SceneItemTransformMessage, error) {
	var sceneTransform types.SceneItemTransformMessage
	err := json.Unmarshal(jsonStr, &sceneTransform)
	if err != nil {
		return nil, err
	}
	return &sceneTransform, nil
}

//
// Sending OUT to the webserver
/////////////////////////////////

func (ctl *ObsController) SendPing() error {
	return ctl.WebClient.SendAction("ping", []byte("{}"))
}

func (ctl *ObsController) SendSceneItemsToServer() error {
	_, sceneItems, err := ctl.GetSelectedSceneItems()
	if err != nil {
		return err
	}

	dataWrappers := make([]types.DataSceneItemDetails, 0)
	for _, item := range sceneItems {
		itemData := types.SceneItemDetails{
			ID:     item.SceneItemId,
			X:      item.XLocation,
			Y:      item.YLocation,
			Width:  fmt.Sprintf("%f", item.Width),
			Height: fmt.Sprintf("%f", item.Height),
			Info:   "some data to register later",
			ZIndex: 10,
		}
		dataItemData := types.DataSceneItemDetails{
			Data: []types.SceneItemDetails{itemData},
		}
		dataWrappers = append(dataWrappers, dataItemData)
	}

	jsonPayload, err := json.Marshal(dataWrappers)
	if err != nil {
		return err
	}
	return ctl.WebClient.SendAction("update_scene_items", jsonPayload)
}

func (ctl *ObsController) SendObsSizeConfig(config *config.GetVideoSettingsResponse) error {

	payload := types.ObsSize{
		OutputWidth:  config.BaseWidth,
		OutputHeight: config.BaseHeight,
	}

	jsonPayload, err := json.Marshal(payload)
	log.Printf("Sending OBS Size UserConfig\n")
	if err != nil {
		return err
	}
	return ctl.WebClient.SendAction("update_video_settings", jsonPayload)
}

func (ctl *ObsController) SendWindowConfig() error {
	jsonPayload, err := json.Marshal(ctl.WindowConfig)
	if err != nil {
		return err
	}
	return ctl.WebClient.SendAction("update_bounds", jsonPayload)
}

func (ctl *ObsController) SendInfoWindowConfig() error {
	infoWindowData, err := ReadInfoWindowData("infoWindowDataConfig.json")
	jsonPayload, err := json.Marshal(infoWindowData)
	if err != nil {
		return err
	}
	return ctl.WebClient.SendAction("update_info_window_config", jsonPayload)
}
