package server

type Message struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}
