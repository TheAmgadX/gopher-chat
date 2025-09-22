package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/TheAmgadX/gopher-chat/internal/middleware"
	"github.com/TheAmgadX/gopher-chat/internal/server"
	"github.com/TheAmgadX/gopher-chat/internal/utils"
	"github.com/gorilla/websocket"
)

// TODO: make a design for this Hub ptr lifecycle.
var hub *server.Hub

func init() {
	hub = server.NewHub()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handles POST /login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJsonErrors(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token, err := middleware.GenerateJWT(req.Username)
	if err != nil {
		utils.WriteJsonErrors(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	// Send the JSON response with the token
	utils.WriteJson(w, map[string]any{"token": token})
}

// Handles GET /ws (and is wrapped by AuthMiddleware)
func NewConnectionHandler(w http.ResponseWriter, r *http.Request) {
	// Get user info from the context (placed by AuthMiddleware)
	claims := r.Context().Value(middleware.UserClaimsKey).(*middleware.Claims)

	// Upgrade the connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgrading connection: %v\n", err)
		return
	}

	req := server.OperationsRequest{
		Username: claims.Username,
		RoomName: "",
		Conn:     ws,
	}

	// register the new client in the Hub.
	err = hub.RegisterUser(req)

	if err != nil {
		ws.WriteJSON(map[string]any{
			"type":    "error",
			"message": err.Error(),
		})

		ws.Close()
		return
	}

	log.Printf("user: %v connected\n", claims.Username)
}
