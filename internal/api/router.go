package api

import (
	"net/http"

	"github.com/TheAmgadX/gopher-chat/internal/api/handlers"
	"github.com/TheAmgadX/gopher-chat/internal/middleware"
	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	router := mux.NewRouter()

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

	// create WebSocket endpoint
	protected.HandleFunc("/ws", handlers.NewConnectionHandler).Methods("GET")

	// --- File Server for your UI ---
	staticFileServer := http.FileServer(http.Dir("./web/static/"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticFileServer))

	fileServer := http.FileServer(http.Dir("./web/templates/"))
	router.PathPrefix("/").Handler(http.StripPrefix("/", fileServer))

	return router
}
