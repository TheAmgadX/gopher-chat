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
			// The readPump will likely catch the disconnection first, but this is for safety.
			log.Printf("client: %v has been disconnected", c.UserName)
			break
		}
	}
}

func (c *Client) readPump(hub *Hub) {
	defer func() {
		// This defer block now handles all cleanup.
		if c.Room != nil {
			req := OperationsRequest{
				Username: c.UserName,
			}

			hub.UnregisterUser(req)
		}

		c.Conn.Close()
		log.Printf("client: %v has been disconnected", c.UserName)
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("error: %s\n", err.Error())
		}
		msg := Message{
			Type:     "message",
			Username: c.UserName,
			Content:  string(message),
		}

		c.Room.Broadcast <- &msg
	}
}
