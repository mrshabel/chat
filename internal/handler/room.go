package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/mrshabel/chat/internal/model"
	"github.com/mrshabel/chat/internal/service"
	"github.com/mrshabel/chat/internal/service/ws"
	"github.com/mrshabel/chat/internal/util"
)

type RoomHandler struct {
	Hub            *ws.Hub
	service        *service.RoomService
	userService    *service.UserService
	messageService *service.MessageService
}

func NewRoomHandler(hub *ws.Hub, service *service.RoomService, userService *service.UserService, messageService *service.MessageService) *RoomHandler {
	return &RoomHandler{
		Hub:            hub,
		service:        service,
		userService:    userService,
		messageService: messageService,
	}
}

func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	userID, err := util.GetParamUUID(r, "userId")
	if err != nil {
		util.WriteError(w, "User ID is required", http.StatusUnprocessableEntity)
		return
	}
	roomID, err := util.GetQueryUUID(r, "roomId")
	if err != nil {
		util.WriteError(w, "Invalid room ID", http.StatusUnprocessableEntity)
		return
	}

	// retrieve user details
	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			util.WriteError(w, "User account not found", http.StatusNotFound)
			return
		}
		util.WriteError(w, "Failed to join room", http.StatusInternalServerError)
		return
	}

	// verify that room exists. the in-memory room will be created only when it exists in the db
	room, err := h.service.GetByID(r.Context(), roomID)
	if err != nil {
		if errors.Is(err, service.ErrRoomNotFound) {
			util.WriteError(w, "Room not found", http.StatusNotFound)
			return
		}
		util.WriteError(w, "Failed to join room", http.StatusInternalServerError)
		return
	}
	if inMemRoom := h.Hub.GetRoom(roomID); inMemRoom == nil {
		h.Hub.Rooms[roomID.String()] = &ws.Room{
			Clients:  make(map[string]*ws.Client),
			ID:       roomID,
			Name:     room.Name,
			Messages: make([]*model.Message, 0, ws.MaxMessageLimit),
		}
	}

	// upgrade client http connection to websocket
	conn, err := ws.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade connection %v\n", err)
		util.WriteError(w, "Failed to join room", http.StatusInternalServerError)
		return
	}

	// register client
	client := &ws.Client{
		ID:       user.ID,
		RoomID:   roomID,
		Username: user.Username,
		Hub:      h.Hub,
		Conn:     conn,
		Inbox:    make(chan *model.Message),
	}
	client.Hub.Register <- client

	// handle connection reads and writes
	go client.WritePump()
	client.ReadPump()
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req model.CreateRoomReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteError(w, "Invalid data format", http.StatusUnprocessableEntity)
		return
	}

	room, err := h.service.Create(r.Context(), &req)
	if err != nil {
		log.Println(err)
		util.WriteError(w, "Failed to created room", http.StatusInternalServerError)
		return
	}

	// preload room
	h.Hub.Rooms[room.ID.String()] = &ws.Room{
		ID:       room.ID,
		Name:     room.Name,
		Clients:  make(map[string]*ws.Client),
		Messages: make([]*model.Message, 0),
	}

	util.WriteJSON(w, room, http.StatusCreated)
}

func (h *RoomHandler) GetRoomByID(w http.ResponseWriter, r *http.Request) {
	id, err := util.GetParamUUID(r, "id")
	if err != nil {
		util.WriteError(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	room, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrRoomNotFound) {
			util.WriteError(w, "Room not found", http.StatusNotFound)
			return
		}
		util.WriteError(w, "Failed to retrieve room", http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, room, http.StatusOK)
}

func (h *RoomHandler) GetAllRooms(w http.ResponseWriter, r *http.Request) {
	skip, limit := util.GetPaginationQuery(r, 1, 50)
	rooms, err := h.service.GetAll(r.Context(), limit, skip)
	if err != nil {
		log.Println(err)
		util.WriteError(w, "Failed to retrieve rooms", http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, rooms, http.StatusOK)
}

func (h *RoomHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	roomID, err := util.GetParamUUID(r, "id")
	if err != nil {
		util.WriteError(w, "Invalid room ID", http.StatusBadRequest)
		return
	}
	var req model.CreateRoomMemberReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	member, err := h.service.AddMember(r.Context(), roomID, req.UserID, string(model.Member))
	if err != nil {
		log.Println(err)
		util.WriteError(w, "Failed to add member", http.StatusInternalServerError)
		return
	}
	util.WriteJSON(w, member, http.StatusCreated)
}

func (h *RoomHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	roomID, err := util.GetParamUUID(r, "id")
	if err != nil {
		util.WriteError(w, "Invalid room ID", http.StatusBadRequest)
		return
	}
	skip, limit := util.GetPaginationQuery(r, 1, 50)

	members, err := h.service.GetAllMembers(r.Context(), roomID, limit, skip)
	if err != nil {
		util.WriteError(w, "Failed to retrieve room members", http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, members, http.StatusOK)
}

func (h *RoomHandler) GetActiveRoomMembers(w http.ResponseWriter, r *http.Request) {
	id, err := util.GetParamUUID(r, "id")
	if err != nil {
		util.WriteError(w, "Room ID is required", http.StatusUnprocessableEntity)
		return
	}

	// verify that room exists with active members
	room := h.Hub.GetRoom(id)
	if room == nil {
		util.WriteError(w, "Room not found or has inactive users", http.StatusNotFound)
		return
	}

	var users []model.User
	for _, client := range room.Clients {
		users = append(users, model.User{
			ID:       client.ID,
			Username: client.Username,
		})
	}

	util.WriteJSON(w, users, 200)
}

// GetAllRoomMessages retrieves the most recent messages in the given room
func (h *RoomHandler) GetAllRoomMessages(w http.ResponseWriter, r *http.Request) {
	roomID, err := util.GetParamUUID(r, "id")
	if err != nil {
		util.WriteError(w, "Invalid room ID", http.StatusBadRequest)
		return
	}
	skip, limit := util.GetPaginationQuery(r, 1, 50)

	messages, err := h.messageService.GetByRoomID(r.Context(), roomID, limit, skip)
	if err != nil {
		util.WriteError(w, "Failed to retrieve room messages", http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, messages, http.StatusOK)
}
