package main

import (
	"errors"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var ErrNotConnected = errors.New("websocket is not connected")

type WebSocketClient struct {
	isConnected bool
	url         string
	conn        *websocket.Conn

	connectedCb func()
}

func (ws *WebSocketClient) closeAndReconnect() {
	ws.Close()
	go func() {
		ws.connect()
	}()
}

func (ws *WebSocketClient) Close() {

	if ws.conn != nil {
		ws.conn.Close()
	}
	ws.isConnected = false
}

func Dial(url string, connectedCb func()) *WebSocketClient {

	ws := &WebSocketClient{
		url:         url,
		connectedCb: connectedCb,
	}

	go func() {
		ws.connect()
	}()

	return ws
}

func (ws *WebSocketClient) connect() {
	for {
		var err error
		ws.conn, _, err = websocket.DefaultDialer.Dial(ws.url, nil)
		if err == nil {
			ws.isConnected = true
			ws.connectedCb()
			break
		}
		log.Errorf("Failed to dial WebSocket: %v", err)
		time.Sleep(10 * time.Second)
	}

}

func (ws *WebSocketClient) WriteMessage(messageType int, data []byte) error {
	err := ErrNotConnected
	if ws.IsConnected() {
		err = ws.conn.WriteMessage(messageType, data)
		if err != nil {
			ws.closeAndReconnect()
		}
	}

	return err
}

func (ws *WebSocketClient) ReadMessage() (messageType int, message []byte, err error) {
	err = ErrNotConnected

	if ws.IsConnected() {
		messageType, message, err = ws.conn.ReadMessage()
		if err != nil {
			ws.closeAndReconnect()
		}
	}

	return
}

func (ws *WebSocketClient) IsConnected() bool {
	return ws.isConnected
}
