package server

import (
	"fmt"
	"sync"
)

type Hub struct {
	Clients map[string]*Client
	Rooms   map[string]*Room
	Mux     sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Clients: make(map[string]*Client),
		Rooms:   make(map[string]*Room),
	}
}

func (h *Hub) CreateRoom(request OperationsRequest) (*Room, error) {
	if request.RoomName == "" {
		return nil, fmt.Errorf("error: room name cannot be empty")
	}

	h.Mux.RLock()
	client := h.Clients[request.Username]

	if client == nil {
		h.Mux.RUnlock()
		return nil, fmt.Errorf("error: username not found")
	}

	// if room already exists just no operation.
	if room, ok := h.Rooms[request.RoomName]; ok {
		h.Mux.RUnlock()
		return room, nil
	}

	h.Mux.RUnlock()
	room := newRoom(request.RoomName)

	h.Mux.Lock()
	h.Rooms[request.RoomName] = room
	h.Mux.Unlock()

	room.Register <- client
	client.Room = room
	go room.run(h)

	return room, nil
}

func (h *Hub) JoinRoom(request OperationsRequest) error {
	h.Mux.RLock()

	// check client and room exist.
	client, clientOK := h.Clients[request.Username]
	room, roomOK := h.Rooms[request.RoomName]
	h.Mux.RUnlock()

	if !roomOK {
		return fmt.Errorf("error: room not found")
	}

	if !clientOK {
		return fmt.Errorf("error: username not found")
	}

	// if the user exist in a room unregister him.
	if client.Room != nil {
		client.Room.Unregister <- client
	}

	room.Register <- client

	client.Room = room

	return nil
}

func (h *Hub) LeaveRoom(request OperationsRequest) error {
	h.Mux.RLock()
	// check client and room exist.
	client, clientOK := h.Clients[request.Username]
	h.Mux.RUnlock()

	if client.Room == nil {
		return fmt.Errorf("error: client is not in a room")
	}

	if !clientOK {
		return fmt.Errorf("error: username not found")
	}

	client.Room.Unregister <- client
	client.Room = nil

	return nil
}

func (h *Hub) SendMessage(msg *Message) error {
	if msg.Content == "" {
		return fmt.Errorf("error: message cannot be empty")
	}

	if msg.Username == "" {
		return fmt.Errorf("error: username cannot be empty")
	}

	if msg.Type == "" {
		return fmt.Errorf("error: message type cannot be empty")
	}

	h.Mux.RLock()
	client := h.Clients[msg.Username]
	h.Mux.RUnlock()

	if client == nil {
		return fmt.Errorf("error: user not found")
	}

	if client.Room == nil {
		return fmt.Errorf("error: user %v is not in room", msg.Username)
	}

	client.Room.Broadcast <- msg

	return nil
}

func (h *Hub) ListRooms() []string {
	h.Mux.RLock()
	defer h.Mux.RUnlock()

	if len(h.Rooms) == 0 {
		return nil
	}

	roomsNames := make([]string, 0, len(h.Rooms))

	for name := range h.Rooms {
		roomsNames = append(roomsNames, name)
	}

	return roomsNames
}

func (h *Hub) GetRoomUsers(request OperationsRequest) ([]*Client, error) {
	h.Mux.RLock()
	room := h.Rooms[request.RoomName]
	h.Mux.RUnlock()

	if room == nil {
		return nil, fmt.Errorf("error: room %v not found", request.RoomName)
	}

	return room.getMembers(), nil
}

func (h *Hub) RegisterUser(request OperationsRequest) error {
	if request.Username == "" {
		return fmt.Errorf("error: username cannot be empty")
	}

	h.Mux.RLock()
	if _, found := h.Clients[request.Username]; found {
		return fmt.Errorf("error: username %v already exist", request.Username)
	}

	room := h.Rooms[request.RoomName]

	h.Mux.RUnlock()

	client := &Client{
		UserName: request.Username,
		Send:     make(chan *Message, 250),
		Conn:     request.Conn,
	}

	if room != nil {
		client.Room = room
		room.Register <- client
	} else {
		client.Room = nil
	}

	h.Mux.Lock()

	h.Clients[request.Username] = client

	h.Mux.Unlock()

	go client.writePump()
	go client.readPump(h)

	return nil
}

func (h *Hub) UnregisterUser(request OperationsRequest) error {
	if request.Username == "" {
		return fmt.Errorf("error: username cannot be empty")
	}

	h.Mux.Lock()

	client, found := h.Clients[request.Username]

	if !found {
		return fmt.Errorf("error: username %v not found", request.Username)
	}

	delete(h.Clients, request.Username)

	h.Mux.Unlock()

	close(client.Send) // close the channel to stop the go routine of the client.

	room := client.Room

	if room != nil {
		room.Unregister <- client
		client.Room = nil
	}

	return nil
}

// delete room from the Rooms map
// called by the room itself when it's empty.
func (h *Hub) CloseRoom(name string) {
	h.Mux.Lock()
	defer h.Mux.Unlock()

	if _, ok := h.Rooms[name]; ok {
		delete(h.Rooms, name)
	}
}
