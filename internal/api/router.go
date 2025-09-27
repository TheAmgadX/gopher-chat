package api

import (
	"github.com/TheAmgadX/gopher-chat/internal/api/handlers"
	"github.com/TheAmgadX/gopher-chat/internal/middleware"
	"github.com/TheAmgadX/gopher-chat/internal/server"
	"github.com/gorilla/mux"
)

func NewRouter(hub *server.Hub) *mux.Router {
	router := mux.NewRouter()

	handlers := handlers.APIHandlers{
		Hub: hub,
	}

	// --- Group ALL API endpoints under a single subrouter ---
	// This ensures all of them get the same middleware consistently.
	api := router.PathPrefix("/api").Subrouter()

	// Apply CORS middleware to every API call.
	// You only need to add it once.
	api.Use(middleware.CorsMiddleware)

	// --- PUBLIC API Routes (No Auth Needed) ---
	api.HandleFunc("/login", handlers.LoginHandler).Methods("POST")

	// --- PROTECTED API Routes (Auth Needed) ---
	protected := api.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthMiddleware)

	// create WebSocket endpoint.
	protected.HandleFunc("/ws", handlers.NewConnectionHandler).Methods("GET")

	// rooms list endpoint.
	protected.HandleFunc("/rooms", handlers.GetAllRoomsHandler).Methods("GET")

	// join room endpoint.
	protected.HandleFunc("/join", handlers.JoinRoomHandler).Methods("POST")

	// leave room endpoint.
	protected.HandleFunc("/leave", handlers.LeaveRoomHandler).Methods("POST")

	return router
}
