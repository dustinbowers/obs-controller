package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"obs-controller/controller/types"
)

type WebClient struct {
	Conn    *websocket.Conn
	Message chan string
	Close   chan string
}

func NewWebClient(wsAddr string) (*WebClient, error) {

	// Connect to websocket lobby
	//u := url.URL{Scheme: "ws", Host: wsHost, Path: wsPath}
	conn, _, err := websocket.DefaultDialer.Dial(wsAddr, nil)
	if err != nil {
		return nil, fmt.Errorf("dial web client: %v", err)
	}
	newClient := &WebClient{
		Conn:    conn,
		Message: make(chan string, 10),
		Close:   make(chan string),
	}
	return newClient, nil
}

func (c *WebClient) StartReadPump() {
	for {
		_, message, err := c.Conn.ReadMessage()
		// Report an issue if the server is gone
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Close <- fmt.Sprintf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, []byte("\n"), []byte(" "), -1))
		c.Message <- string(message)
	}
}

// Deprecated: Send is deprecated. Use SendAction instead
//func (c *WebClient) Send(message []byte) error {
//	log.Printf("\tPAYLOAD SENT: %s", message)
//	err := c.Conn.WriteMessage(websocket.TextMessage, message)
//	return err
//}

func (c *WebClient) SendAction(action string, data []byte) error {
	envelope := types.ActionEnvelope{
		Action: action,
		Data:   data,
	}
	jsonPayload, err := json.Marshal(envelope)
	if err != nil {
		log.Printf("Error marshalling envelope: %v", err)
		return err
	}
	log.Printf("OUTBOUND message to WebClient: \n>>>>>>>\t\t%s", string(jsonPayload))
	return c.Conn.WriteMessage(websocket.TextMessage, jsonPayload)
}
