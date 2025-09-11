package server

type commandId int

const (
	CMD_USER_NAME commandId = iota
	CMD_JOIN
	CMD_ROOMS
	CMD_MSG
	CMD_QUIT
)

type Command struct {
	Id     commandId
	Client *Client
	Args   []string
}
