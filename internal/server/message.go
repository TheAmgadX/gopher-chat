package server

type Message struct {
	Type     string `json:"type"` // error or message.
	Username string `json:"username"`
	Content  string `json:"content"`
}
