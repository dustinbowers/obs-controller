package controller

import (
	"bytes"
	"context"
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

func (c *WebClient) Disconnect() error {
	return c.Conn.Close()
}

func (c *WebClient) StartReadPump(ctx context.Context) error {
	defer c.Conn.Close()
	for {
		if ctx.Err() != nil { // when the ctx.Done() channel is 'done', ctx.Err() will not be nil
			return nil
		}
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
	return nil
}

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
	log.Printf("OUTBOUND message to WebClient: %s", string(jsonPayload))
	return c.Conn.WriteMessage(websocket.TextMessage, jsonPayload)
}
