package main

import (
	"log"
	"net"
	"net/http"

	"github.com/TheAmgadX/gopher-chat/internal/api"
	"github.com/TheAmgadX/gopher-chat/internal/server"
)

func main() {
	chatServer := server.InitServer()

	// Start the chat server in a goroutine
	go chatServer.Run()

	// Start TCP server for traditional chat clients
	go func() {
		listener, err := net.Listen("tcp", ":8081")
		if err != nil {
			log.Fatalf("Error starting TCP server: %v", err)
		}
		defer listener.Close()

		log.Println("TCP chat server started on :8081")

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error accepting TCP connection: %v", err)
				continue
			}

			go chatServer.NewClient(conn)
		}
	}()

	api.Init(chatServer)

	// Setup HTTP router
	router := api.NewRouter(chatServer)

	log.Println("Web server starting on :8080")
	log.Println("TCP chat server on :8081")
	log.Println("WebSocket endpoint: ws://localhost:8080/ws?token=<your_jwt_token>")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Error starting web server: %v", err)
	}
}
