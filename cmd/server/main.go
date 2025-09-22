package main

import (
	"log"
	"net/http"

	"github.com/TheAmgadX/gopher-chat/internal/api"
)

func main() {
	// Setup HTTP router
	router := api.NewRouter()

	log.Println("Web server starting on :8000")

	if err := http.ListenAndServe(":8000", router); err != nil {
		log.Fatalf("Error starting web server: %v", err)
	}
}
