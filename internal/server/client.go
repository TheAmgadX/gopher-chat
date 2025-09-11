package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type Client struct {
	Conn     net.Conn
	UserName string
	Room     *Room
	Commands chan<- Command
}

func (c *Client) Err(err error) {
	c.Conn.Write([]byte("ERROR: " + err.Error() + "\n"))
}

func (c *Client) Msg(msg string) {
	c.Conn.Write([]byte("==> " + msg + "\n"))
}

func (c *Client) ReadInput() {
	for {
		msg, err := bufio.NewReader(c.Conn).ReadString('\n')

		if err != nil {
			log.Printf("Error: unable to read the message: %s", err.Error())
			return
		}

		msg = strings.Trim(msg, "\r\n")

		args := strings.Split(msg, " ")

		cmd := strings.TrimSpace(args[0])

		switch cmd {
		case "/username":
			c.Commands <- Command{
				Id:     CMD_USER_NAME,
				Client: c,
				Args:   args,
			}

		case "/join":
			c.Commands <- Command{
				Id:     CMD_JOIN,
				Client: c,
				Args:   args,
			}

		case "/rooms":
			c.Commands <- Command{
				Id:     CMD_ROOMS,
				Client: c,
				Args:   args,
			}

		case "/msg":
			c.Commands <- Command{
				Id:     CMD_MSG,
				Client: c,
				Args:   args,
			}

		case "/quit":
			c.Commands <- Command{
				Id:     CMD_QUIT,
				Client: c,
				Args:   args,
			}

		default:
			c.Err(fmt.Errorf("unknown command: %s", cmd))
		}
	}
}
