package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mrshabel/chat/internal/model"
	"github.com/mrshabel/chat/internal/service/ws"
	"github.com/mrshabel/chat/internal/util"
)

type RoomHandler struct {
	Hub *ws.Hub
}

func NewRoomHandler(hub *ws.Hub) *RoomHandler {
	return &RoomHandler{
		Hub: hub,
	}
}

func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userId"]
	if userID == "" {
		util.WriteError(w, "User ID is required", http.StatusUnprocessableEntity)
		return
	}
	q := r.URL.Query()
	roomID := q.Get("roomId")
	username := q.Get("username")
	if roomID == "" || username == "" {
		util.WriteError(w, "Room ID and Username is required", http.StatusUnprocessableEntity)
		return
	}

	// verify that room exists. the in-memory room will be created only when it exists in the db
	if h.Hub.GetRoom(roomID) == nil {
		log.Printf("Room %v not found\n", roomID)
		util.WriteError(w, "Room not found", http.StatusNotFound)
		return
	}

	client := &ws.Client{
		ID:       userID,
		RoomID:   roomID,
		Username: username,
	}

	ws.ServeWS(h.Hub, client, w, r)
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req model.CreateRoomReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteError(w, "Invalid data format", http.StatusUnprocessableEntity)
		return
	}
	if err := req.Validate(); err != nil {
		util.WriteError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// TODO: persist room
	id := uuid.New()
	room := &ws.Room{
		ID:       id.String(),
		Name:     req.Name,
		Clients:  make(map[string]*ws.Client),
		Messages: make([]*model.Message, 0),
	}
	h.Hub.Rooms[id.String()] = room

	res := model.Room{
		ID:        room.ID,
		Name:      room.Name,
		CreatedAt: time.Now().UTC(),
	}

	util.WriteJSON(w, res, 201)
}

func (h *RoomHandler) GetRooms(w http.ResponseWriter, r *http.Request) {
	var rooms []*model.Room
	// TODO: fetch from db and invalidate in-memory clients
	for _, room := range h.Hub.Rooms {
		rooms = append(rooms, &model.Room{
			ID:        room.ID,
			Name:      room.Name,
			CreatedAt: time.Now(),
		})
	}

	util.WriteJSON(w, rooms, 200)
}

func (h *RoomHandler) GetActiveRoomMembers(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		util.WriteError(w, "Room ID is required", http.StatusUnprocessableEntity)
		return
	}
	// verify that room exists
	room := h.Hub.GetRoom(id)
	if room == nil {
		util.WriteError(w, "Room not found", http.StatusNotFound)
		return
	}

	var clients []*model.Client
	for _, client := range room.Clients {
		clients = append(clients, &model.Client{
			ID:       client.ID,
			Username: client.Username,
		})
	}

	util.WriteJSON(w, clients, 200)
}
