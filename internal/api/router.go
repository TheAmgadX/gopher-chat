package api

import (
	"net/http"

	"github.com/TheAmgadX/gopher-chat/internal/middleware"
	"github.com/TheAmgadX/gopher-chat/internal/server"
	"github.com/gorilla/mux"
)

func NewRouter(chatServer *server.Server) *mux.Router {
	router := mux.NewRouter()

	// Auth endpoint
	router.HandleFunc("/api/login", LoginHandler()).Methods("POST")

	// WebSocket endpoint
	router.HandleFunc("/ws", WebSocketHandler(chatServer)).Methods("GET")

	// REST API endpoints for room management.
	api := router.PathPrefix("/api").Subrouter()

	api.HandleFunc("/rooms/", GetRoomsHandler(chatServer)).Methods("GET")
	api.HandleFunc("/user/rename", RenameUserHandler()).Methods("POST")

	api.Use(middleware.CorsMiddleware)

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	return router
}
