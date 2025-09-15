package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/TheAmgadX/gopher-chat/internal/auth"
	"github.com/TheAmgadX/gopher-chat/internal/server"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var hub = struct {
	clients map[*websocket.Conn]*server.WebSocketClient
	sync.Mutex
}{
	clients: make(map[*websocket.Conn]*server.WebSocketClient),
}

type LoginRequest struct {
	Username string `json:"username"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// Start a websocket broadcast goroutine.
func Init(server *server.Server) {
	go func() {
		for msg := range server.WebBroadcastChan {
			BroadCastToWebsockets(msg)
		}
	}()
}

func BroadCastToWebsockets(msg server.WebSocketMessage) {
	hub.Lock()
	defer hub.Unlock()

	chatMsg := server.WebSocketMessage{
		Type:     msg.Type,
		Username: msg.Username,
		Message:  msg.Message,
		Room:     msg.Room,
	}

	if msg.Room != "" {
		for client := range hub.clients {
			if err := client.WriteJSON(chatMsg); err != nil {
				log.Printf("error: writing json to websocket client: %v", err.Error())
				client.Close()
				delete(hub.clients, client)
			}
		}
	}
}

func LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if req.Username == "" {
			http.Error(w, "Invalid username, username cannot be empty.", http.StatusBadRequest)
			return
		}

		token, err := auth.GenerateJWT(req.Username)

		if err != nil {
			http.Error(w, "Couldn't generate token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LoginResponse{Token: token})
	}
}

func GetRoomsHandler(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Mux.Lock()
		rooms := make([]server.RoomInfo, 0, len(s.Rooms))

		for roomName, room := range s.Rooms {
			room.Mux.Lock()
			members := make([]string, 0, len(room.Members))

			for _, client := range room.Members {
				members = append(members, client.UserName)
			}

			hub.Lock()

			for _, wsClient := range hub.clients {
				if wsClient.Room == roomName {
					members = append(members, wsClient.Username)
				}
			}

			hub.Unlock()

			rooms = append(rooms, server.RoomInfo{
				Name:        roomName,
				MemberCount: len(members),
				Members:     members,
			})

			room.Mux.RUnlock()
		}

		s.Mux.RUnlock()

		w.Header().Set("Content-Tyoe", "application/json")
		json.NewEncoder(w).Encode(server.RoomListResponse{Rooms: rooms})
	}
}

func WebSocketHandler(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.URL.Query().Get("token")

		if tokenStr == "" {
			http.Error(w, "Missing auth token", http.StatusUnauthorized)
			return
		}

		claims, err := auth.ValidJWT(tokenStr)
		if err != nil {
			http.Error(w, "Invalid JWT token", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			log.Printf("Websocket upgrade error: %v\n", err.Error())
			return
		}

		hub.Lock()
		hub.clients[conn] = true
		hub.Unlock()

		log.Printf("User %s connected via WebSocket\n", claims.Username)

		s.WebMessages <- server.WebMessage{
			Type:     server.UserJoined,
			Username: claims.Username,
		}

		handleWebSocketMessages(conn, claims.Username, s)
	}
}

func handleWebSocketMessages(conn *websocket.Conn, username string, s *server.Server) {
	defer func() {
		// Unregister client
		hub.Lock()
		delete(hub.clients, conn)
		hub.Unlock()

		conn.Close()
		log.Printf("User %s disconnecte\n", username)

		s.WebMessages <- server.WebMessage{
			Type:     server.UserLeft,
			Username: username,
		}
	}()

	for {
		var msg ChatMessage

		// Read Message from the browser.
		err := conn.ReadJSON(&msg)

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading json from websocket: %v", err)
			}
			break
		}

		// Send the message to the server
		s.WebMessages <- server.WebMessage{
			Type:     server.NewMessage,
			Username: username,
			Message:  msg.Message,
		}
	}
}
