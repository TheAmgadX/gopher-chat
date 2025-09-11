package server

import "net"

type Room struct {
	Name    string
	Members map[net.Addr]*Client
}

func (r *Room) Broadcast(c *Client, msg string) {
	for addr, member := range r.Members {
		if addr != c.Conn.RemoteAddr() {
			member.Msg(msg)
		}
	}
}
