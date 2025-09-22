package server

import "github.com/gorilla/websocket"

type OperationsRequest struct {
	Type     string `json:"type"`
	RoomName string `json:"room"`
	Username string `json:"username"`
	Conn     *websocket.Conn
}
