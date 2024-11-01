package controller

import (
	"encoding/json"
	"fmt"
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/config"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"github.com/andreykaipov/goobs/api/typedefs"
	"log"
	"obs-controller/controller/types"
	"slices"
	"time"
)

// ObsController holds the OBS and Web proxy connections
type ObsController struct {
	ObsClient    *goobs.Client
	WebClient    *WebClient
	WindowConfig *types.WindowConfig
}

func NewController(obsHost string, obsPassword string, webUserId string) (*ObsController, error) {
	// Load WindowConfig
	windowConfig, err := ReadWindowConfig("windowConfig.json")
	if err != nil {
		return nil, fmt.Errorf("read window config err: %v", err)
	}

	// Connect to OBS
	log.Printf("Connecting to OBS...")
	obsClient, err := goobs.New(obsHost, goobs.WithPassword(obsPassword))
	if err != nil {
		return nil, err
	}
	log.Println("Done")
	if err := PrintObsVersion(obsClient); err != nil {
		return nil, fmt.Errorf("error printing obs version: %v", err)
	}

	// Fetch room key for user
	log.Printf("Fetching room key for %s\n", webUserId)
	newRoomUrl := fmt.Sprintf("https://websocket.matissetec.dev/lobby/new?user=%s", webUserId)
	roomKey, err := GetRoomKey(newRoomUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to get room key for user %s: %v", webUserId, err)
	}
	log.Printf("\tRoom key: %s\n", roomKey)

	// Connect to Websocket Proxy
	log.Printf("Connecting to websocket proxy...")
	wsAddr := fmt.Sprintf("wss://websocket.matissetec.dev/lobby/connect/streamer?user=%s&key=%s", webUserId, roomKey)
	webClient, err := NewWebClient(wsAddr)
	if err != nil {
		return nil, fmt.Errorf("connection to websocket proxy failed: %v", err)
	}
	log.Printf("Done\n")

	newClient := ObsController{
		ObsClient:    obsClient,
		WebClient:    webClient,
		WindowConfig: windowConfig,
	}
	return &newClient, nil
}

func (ctl *ObsController) Cleanup() error {
	return ctl.ObsClient.Disconnect()
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

func (ctl *ObsController) Run() error {
	log.Printf("Starting Webclient read pump...")

	// All this function does is listen for incoming web proxy messages, and sends them
	// into a buffered channel that we can process later in this function
	// NOTE: this is a go-routine, so this function continues running in its own thread
	go ctl.WebClient.StartReadPump()
	log.Printf("Running")

	//sceneItemIDs, sceneItems, err := ctl.GetSelectedSceneItems()
	//if err != nil {
	//	return err
	//}
	//fmt.Printf("Scene items list: %#+v\n", sceneItems)

	videoOutputSettings, err := ctl.GetVideoOutputSettings()
	if err != nil {
		return err
	}

	//// Send list of window IDs
	//if err := ctl.SendWindowIDs(sceneItemIDs); err != nil {
	//	return err
	//}

	// Send a "ping" to say hello
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
		case msg := <-ctl.WebClient.Close:
			log.Printf("Websocket proxy closed: %v", msg)
			return nil
		case message := <-ctl.WebClient.Message:
			log.Printf("Websocket proxy message:\n<<<<<<<\t\t%v", message)

			var action types.ActionEnvelope
			err := json.Unmarshal([]byte(message), &action)
			if err != nil {
				log.Printf("error unmarshalling message: %v", err)
				break
			}
			log.Printf("Action unmarshaled: %#+v", action)

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
				log.Printf("PARSED sceneItemTransformCommand: %#+v", sceneItemTransformCommand)

				//sceneItemID, err := strconv.Atoi(sceneItemTransformCommand.ItemID)
				//if err != nil {
				//	log.Printf("sceneItemID conversion error: %v", err)
				//	break
				//}

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

			// If server says hello, send the "welcome" package
			//if message == "Hello Server!" {
			//
			//	// Send video output settings
			//	if err := ctl.SendObsSizeConfig(videoOutputSettings); err != nil {
			//		return err
			//	}
			//
			//	// Send window config
			//	if err := ctl.SendWindowConfig(); err != nil {
			//		return err
			//	}
			//
			//	// Send info window data config
			//	if err := ctl.SendInfoWindowConfig(); err != nil {
			//		return err
			//	}
			//
			//	// Send the new scene item transforms back to the web proxy
			//	err := ctl.SendSceneItemsToServer()
			//	if err != nil {
			//		return err
			//	}
			//
			//} else {
			//	// If we receive a json payload, let's try to parse it
			//	if sceneItemTransformCommand, err := ctl.ParseSceneItemTransform(message); err == nil {
			//		sceneItemID, err := strconv.Atoi(sceneItemTransformCommand.ItemID)
			//		if err != nil {
			//			break
			//		}
			//
			//		// Get updated video settings
			//		videoOutputSettings, err := ctl.GetVideoOutputSettings()
			//		if err != nil {
			//			return err
			//		}
			//
			//		// Get Scene Item Transform details
			//		currentSceneItemTransform, err := ctl.GetSceneItemTransformByID("Scene", sceneItemID)
			//		if err != nil {
			//			return err
			//		}
			//
			//		currentSceneItemTransform.PositionX = sceneItemTransformCommand.X * videoOutputSettings.OutputWidth
			//		currentSceneItemTransform.PositionY = sceneItemTransformCommand.Y * videoOutputSettings.OutputHeight
			//		currentSceneItemTransform.BoundsWidth = 1.0  // TODO: Not sure why these are necessary
			//		currentSceneItemTransform.BoundsHeight = 1.0 // 	  but they are...
			//
			//		// Send the newly received Scene item transform to OBS
			//		err = ctl.TransformSceneItemByID(
			//			"Scene",
			//			sceneItemID,
			//			currentSceneItemTransform)
			//		if err != nil {
			//			return err
			//		}
			//
			//		// After the transforms are updated, we need to update the web proxy with the new coordinates
			//		err = ctl.SendSceneItemsToServer()
			//		if err != nil {
			//			return err
			//		}
			//	} else {
			//		log.Printf("Attempting to ParseSceneItemTransform failed: %v", err)
			//	}
			//}
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
	fmt.Printf("SceneItemTransform: %#+v\n", newSceneItemTransform)
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
	//dataWrappersContainer := types.DataDataSceneItemData{
	//	Data: dataWrappers,
	//}

	jsonPayload, err := json.Marshal(dataWrappers)
	if err != nil {
		return err
	}
	return ctl.WebClient.SendAction("update_scene_items", jsonPayload)
}

//func (ctl *ObsController) SendWindowIDs(sceneItemIDs []int) error {
//	for _, id := range sceneItemIDs {
//		params := types.GetPositionsParams{
//			Command: "get_positions",
//			ID:      id,
//		}
//		jsonPayload, err := json.Marshal(params)
//		if err != nil {
//			return fmt.Errorf("json marshalling failed: %v", err)
//		}
//
//		if err = ctl.WebClient.Send(jsonPayload); err != nil {
//			return fmt.Errorf("sending json payload failed: %v", err)
//		}
//	}
//	return nil
//}

func (ctl *ObsController) SendObsSizeConfig(config *config.GetVideoSettingsResponse) error {
	//payload := types.ObsSizeContainer{
	//	ObsSize: types.ObsSize{
	//		OutputWidth:  config.OutputWidth,
	//		OutputHeight: config.OutputHeight,
	//	},
	//}

	payload := types.ObsSize{
		OutputWidth:  config.BaseWidth,
		OutputHeight: config.BaseHeight,
	}

	jsonPayload, err := json.Marshal(payload)
	log.Printf("Sending OBS Size Config\n")
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
