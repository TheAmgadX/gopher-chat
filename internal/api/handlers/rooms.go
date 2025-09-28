package handlers

import (
	"net/http"

	"github.com/TheAmgadX/gopher-chat/internal/middleware"
	"github.com/TheAmgadX/gopher-chat/internal/server"
	"github.com/TheAmgadX/gopher-chat/internal/utils"
)

func (h *APIHandlers) GetAllRoomsHandler(w http.ResponseWriter, r *http.Request) {
	rooms := h.Hub.ListRooms()

	err := utils.WriteJson(w, map[string]any{
		"rooms": rooms,
	})

	if err != nil {
		utils.WriteJsonErrors(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *APIHandlers) JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserClaimsKey).(*middleware.Claims)

	username := claims.Username
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

	err := h.Hub.JoinRoom(req)

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

func (h *APIHandlers) LeaveRoomHandler(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserClaimsKey).(*middleware.Claims)

	username := claims.Username

	if username == "" {
		utils.WriteJsonErrors(w, "username is required", http.StatusBadRequest)
		return
	}

	req := server.OperationsRequest{
		Username: username,
	}

	err := h.Hub.LeaveRoom(req)

	if err != nil {
		utils.WriteJsonErrors(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = utils.WriteJson(w, map[string]any{
		"status": "ok",
	})

	if err != nil {
		utils.WriteJsonErrors(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *APIHandlers) CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.UserClaimsKey).(*middleware.Claims)

	username := claims.Username
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

	newRoom, err := h.Hub.CreateRoom(req)

	if err != nil {
		utils.WriteJsonErrors(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = utils.WriteJson(w, map[string]any{
		"status":   "ok",
		"username": username,
		"room":     newRoom.Name,
	})

	if err != nil {
		utils.WriteJsonErrors(w, err.Error(), http.StatusInternalServerError)
	}
}
