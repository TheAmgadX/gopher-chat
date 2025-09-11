package main

import (
	"log"
	"net"

	"github.com/TheAmgadX/gopher-chat/internal/server"
)

func main() {
	sv := server.InitServer()

	go sv.Run()

	listener, err := net.Listen("tcp", ":8888")

	if err != nil {
		log.Fatalf("Error, unable to start server: %s", err.Error())
	}
	defer listener.Close()

	log.Printf("started gopher chat server on :8888")

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Printf("unable to accept connection, %s", err.Error())
			continue
		}

		go sv.NewClient(conn)
	}
}
