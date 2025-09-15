package server

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn     *websocket.Conn
	UserName string
	Room     *Room
	Send     chan *Message
	Mux      sync.RWMutex
}
