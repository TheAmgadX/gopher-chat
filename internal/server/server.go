package server

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type Server struct {
	Rooms    map[string]*Room
	Commands chan Command
}

func InitServer() *Server {
	return &Server{
		Rooms:    make(map[string]*Room),
		Commands: make(chan Command),
	}
}

func (s *Server) RenameUser(c *Client, args []string) {
	c.UserName = args[1]
	c.Msg(fmt.Sprintf("new user name: %s", c.UserName))
}

func (s *Server) Join(c *Client, args []string) {
	roomName := args[1]

	room, ok := s.Rooms[roomName]

	if !ok {
		room = &Room{
			Name:    roomName,
			Members: make(map[net.Addr]*Client),
		}

		s.Rooms[roomName] = room
	}

	room.Members[c.Conn.RemoteAddr()] = c

	s.quitCurrentRoom(c)

	c.Room = room

	// notify all current users in the room.
	room.Broadcast(c, fmt.Sprintf("%s has joined the room.", c.UserName))

	c.Msg(fmt.Sprintf("welcome to %s", room.Name))
}

func (s *Server) RoomsList(c *Client) {
	var rooms []string

	for name := range s.Rooms {
		rooms = append(rooms, name)
	}

	c.Msg(fmt.Sprintf("available rooms: $s", strings.Join(rooms, ", ")))
}

func (s *Server) Message(c *Client, args []string) {
	if c.Room == nil {
		c.Err(fmt.Errorf("you must join a room first to send messages"))
		return
	}

	c.Room.Broadcast(c, c.UserName+" => "+strings.Join(args[1:], " "))
}

func (s *Server) Quit(c *Client, args []string) {
	log.Printf("client has disconnected: %s", c.UserName)

	s.quitCurrentRoom(c)

	c.Msg("you left the room :(")
	c.Conn.Close()
}

func (s *Server) quitCurrentRoom(c *Client) {
	// if the user was in a room:
	if c.Room != nil {
		delete(c.Room.Members, c.Conn.RemoteAddr())

		// notify room members that user left the room:
		c.Room.Broadcast(c, fmt.Sprintf("%s has left the room.", c.UserName))
	}
}

func (s *Server) Run() {
	for cmd := range s.Commands {
		switch cmd.Id {
		case CMD_USER_NAME:
			s.RenameUser(cmd.Client, cmd.Args)
		case CMD_JOIN:
			s.Join(cmd.Client, cmd.Args)
		case CMD_ROOMS:
			s.RoomsList(cmd.Client)
		case CMD_MSG:
			s.Message(cmd.Client, cmd.Args)
		case CMD_QUIT:
			s.Quit(cmd.Client, cmd.Args)
		}
	}
}

func (s *Server) NewClient(conn net.Conn) {
	log.Printf("new client has connected: $s", conn.RemoteAddr().String())

	c := &Client{
		Conn:     conn,
		UserName: "annonymous",
		Commands: s.Commands,
	}

	c.ReadInput()
}
