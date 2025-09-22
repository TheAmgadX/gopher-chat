package handlers

import (
	"net/http"

	"github.com/TheAmgadX/gopher-chat/internal/server"
	"github.com/TheAmgadX/gopher-chat/internal/utils"
)

func GetAllRoomsHandler(w http.ResponseWriter, r *http.Request) {
	rooms := hub.ListRooms()

	err := utils.WriteJson(w, map[string]any{
		"rooms": rooms,
	})

	if err != nil {
		utils.WriteJsonErrors(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	room := r.URL.Query().Get("room")

	if username == "" {
		utils.WriteJsonErrors(w, "username is required", http.StatusBadRequest)
		return
	}

	if room == "" {
		utils.WriteJsonErrors(w, "room name is required", http.StatusBadRequest)
		return
	}

	req := server.OperationsRequest{
		Username: username,
		RoomName: room,
	}

	err := hub.JoinRoom(req)

	if err != nil {
		utils.WriteJsonErrors(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = utils.WriteJson(w, map[string]any{
		"status":   "ok",
		"username": username,
		"room":     room,
	})

	if err != nil {
		utils.WriteJsonErrors(w, err.Error(), http.StatusInternalServerError)
	}
}
