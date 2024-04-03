package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

type WsConnection struct {
	w  *websocket.Conn
	mu sync.Mutex
}

func NewWsConnection(w *websocket.Conn) *WsConnection {
	return &WsConnection{w: w}
}

func (w *WsConnection) Send(message any) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.w.WriteJSON(message)
}
