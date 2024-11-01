package types

import "encoding/json"

type ActionEnvelope struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type WindowDetails struct {
	SceneItemId int     `json:"windowId"`
	SourceName  string  `json:"windowName"`
	Width       float64 `json:"width"`
	Height      float64 `json:"height"`
	XLocation   float64 `json:"xLocation"`
	YLocation   float64 `json:"yLocation"`
}

type GetPositionsParams struct {
	Command string `json:"command"`
	ID      int    `json:"id"`
}

type ObsSizeContainer struct {
	ObsSize ObsSize `json:"obsSize"`
}

type ObsSize struct {
	OutputWidth  float64 `json:"output_width"`
	OutputHeight float64 `json:"output_height"`
}

type SceneItemDetails struct {
	ID     int     `json:"name"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  string  `json:"width"`
	Height string  `json:"height"`
	Info   string  `json:"info"`
	ZIndex int     `json:"zIndex"`
}

type DataSceneItemDetails struct {
	Data []SceneItemDetails `json:"data"`
}

type DataDataSceneItemData struct {
	Data []DataSceneItemDetails `json:"data"`
}

type SceneItemTransformMessage struct {
	Color  string  `json:"color"`
	ItemID int     `json:"id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	UserID string  `json:"user_id"`
}
