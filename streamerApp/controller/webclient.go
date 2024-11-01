package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"obs-controller/controller/types"
	"time"
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
			log.Printf("ReadPump shutting down...")
			c.GracefulClose()
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

func (c *WebClient) GracefulClose() error {
	// Send a WebSocket close message
	deadline := time.Now().Add(time.Minute)
	err := c.Conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		deadline,
	)
	if err != nil {
		return err
	}

	// Set deadline for reading the next message
	err = c.Conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return err
	}
	// Read messages until the close message is confirmed
	for {
		_, _, err = c.Conn.NextReader()
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			break
		}
		if err != nil {
			break
		}
	}
	// Close the TCP connection
	err = c.Conn.Close()
	if err != nil {
		return err
	}
	return nil
}
