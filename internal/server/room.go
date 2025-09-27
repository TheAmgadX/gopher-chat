package server

import (
	"log"
	"sync"
)

type Room struct {
	Name       string
	Broadcast  chan *Message
	Register   chan *Client
	Unregister chan *Client
	Members    map[*Client]bool
	Mux        sync.RWMutex
}

func newRoom(name string) *Room {
	return &Room{
		Name:       name,
		Broadcast:  make(chan *Message, 1024),
		Register:   make(chan *Client, 16),
		Unregister: make(chan *Client, 16),
		Members:    make(map[*Client]bool),
	}
}

func (r *Room) getMembers() []*Client {
	r.Mux.RLock()
	defer r.Mux.RUnlock()

	members := make([]*Client, 0, len(r.Members))

	for member, _ := range r.Members {
		members = append(members, member)
	}

	return members
}

func (r *Room) broadcast(msg *Message) {
	for client := range r.Members {
		select {
		case client.Send <- msg:
		default:
			log.Printf("client %s's send channel is full. Dropping message.", client.UserName)
		}
	}
}

func (r *Room) join(client *Client) {
	// if client exists just skip.
	if _, ok := r.Members[client]; ok {
		return
	}

	r.Members[client] = true
}

func (r *Room) leave(client *Client) {
	// if client doesn't exist just skip.
	if _, ok := r.Members[client]; !ok {
		return
	}

	delete(r.Members, client)
}

func (r *Room) currentMembersCount() int {
	r.Mux.RLock()
	defer r.Mux.RUnlock()
	return len(r.Members)
}

func (r *Room) run(h *Hub) {
	for {
		select {
		case client := <-r.Register:
			r.join(client)

		case client := <-r.Unregister:
			r.leave(client)

			if r.currentMembersCount() == 0 {
				h.CloseRoom(r.Name)
				return
			}

		case msg := <-r.Broadcast:
			r.broadcast(msg)
		}
	}
}
