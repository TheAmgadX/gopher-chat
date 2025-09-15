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

func (h *Hub) CreateRoom(request OperationsRequest) (*Room, error) {
	if request.RoomName == "" {
		return nil, fmt.Errorf("error: room name cannot be empty")
	}

	h.Mux.RLock() // allow read operations.

	// if room already exists just no operation.
	if room, ok := h.Rooms[request.RoomName]; ok {
		h.Mux.RUnlock()
		return room, nil
	}

	h.Mux.RUnlock()
	room := &Room{
		Name:       request.RoomName,
		Members:    make(map[*Client]bool),
		Register:   make(chan *Client, 10),
		Unregister: make(chan *Client, 10),
		Broadcast:  make(chan *Message, 1000),
	}

	h.Mux.Lock()
	h.Rooms[request.RoomName] = room
	h.Mux.Unlock()

	h.Mux.RLock()
	client := h.Clients[request.Username]
	h.Mux.RUnlock()

	if client != nil {
		room.Register <- client
	}

	go room.Run()

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

	room.Register <- client

	return nil
}

func (h *Hub) LeaveRoom(request OperationsRequest) error {
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

	client.Room = nil

	room.Unregister <- client

	return nil
}

func (h *Hub) SendMessage(msg *Message) error {
	if msg.Content == "" {
		return fmt.Errorf("error: message cannot be empty")
	}

	if msg.Username == "" {
		return fmt.Errorf("error: username cannot be empty")
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

func (h *Hub) ListRooms() []*Room {
	h.Mux.RLock()
	defer h.Mux.RUnlock()

	if len(h.Rooms) == 0 {
		return nil
	}

	rooms := make([]*Room, 0, len(h.Rooms))

	for _, room := range h.Rooms {
		rooms = append(rooms, room)
	}

	return rooms
}

func (h *Hub) GetRoomUsers(request OperationsRequest) ([]*Client, error) {
	h.Mux.RLock()
	room := h.Rooms[request.RoomName]
	h.Mux.RUnlock()

	if room == nil {
		return nil, fmt.Errorf("error: room %v not found", request.RoomName)
	}

	return room.GetMembers(), nil
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
		Send:     make(chan *Message),
		Conn:     request.conn,
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

	room := client.Room

	if room != nil {
		room.Unregister <- client
	}

	return nil
}
