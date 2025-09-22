package server

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn     *websocket.Conn
	UserName string
	Room     *Room
	Send     chan *Message
}

func (c *Client) writePump() {
	defer c.Conn.Close()

	for msg := range c.Send {
		data, err := json.Marshal(msg)
		if err != nil {
			log.Printf("msg: %v\n-----\nfrom user: %v\n-----\nTo user: %v\n-----\nis lost\n", msg.Content, msg.Username, c.UserName)
			continue
		}

		err = c.Conn.WriteMessage(websocket.TextMessage, data)

		if err != nil {
			log.Printf("client: %v has been disconnected", c.UserName)
			break
		}
	}
}
