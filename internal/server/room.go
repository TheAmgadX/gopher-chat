package server

import "sync"

type Room struct {
	Name       string
	Broadcast  chan *Message
	Register   chan *Client
	Unregister chan *Client
	Members    map[*Client]bool
	Mux        sync.RWMutex
}

func (r *Room) GetMembers() []*Client {
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
			// drop if client is slow/unresponsive
			delete(r.Members, client) // remove the client from the room.
			client.Room = nil
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

func (r *Room) CurrentMembersCount() int {
	r.Mux.RLock()
	defer r.Mux.RUnlock()
	return len(r.Members)
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.Register:
			r.join(client)

		case client := <-r.Unregister:
			r.leave(client)

			if r.CurrentMembersCount() == 0 {
				return
			}

		case msg := <-r.Broadcast:
			r.broadcast(msg)
		}
	}
}
