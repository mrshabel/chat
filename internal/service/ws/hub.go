package ws

import (
	"context"

	"github.com/google/uuid"
	"github.com/mrshabel/chat/internal/model"
	"github.com/mrshabel/chat/internal/service"
)

// Room holds all connected clients
type Room struct {
	// client id to connection mapping
	Clients  map[string]*Client
	ID       uuid.UUID
	Name     string
	Messages []*model.Message
}

// Hub holds the set of active clients and broadcasts messages to them
type Hub struct {
	// room id to room mapping
	Rooms map[string]*Room

	// inbound messages from clients
	broadcast chan *model.Message

	// register/enter requests from client
	register chan *Client

	// unregister/leave request from client
	unregister chan *Client

	roomService *service.RoomService
}

func NewHub(roomService *service.RoomService) *Hub {
	return &Hub{
		Rooms:       make(map[string]*Room),
		broadcast:   make(chan *model.Message),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		roomService: roomService,
	}
}

// run starts the hub and handles all related events
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// joined specified room and inform members
			room := h.GetRoom(client.RoomID)
			if room == nil {
				// get room from db
				dbRoom, err := h.roomService.GetByID(context.Background(), client.RoomID)
				if err != nil {
					continue
				}

				room = &Room{
					ID:       dbRoom.ID,
					Name:     dbRoom.Name,
					Messages: make([]*model.Message, 0),
					Clients:  make(map[string]*Client),
				}
				h.Rooms[client.RoomID.String()] = room
			}
			room.Clients[client.ID.String()] = client

			// replay messages history to client
			go func() {
				for _, message := range room.Messages {
					client.inbox <- message
				}
			}()

			// TODO: inform room members via notification event

		case client := <-h.unregister:
			// remove client from room and close inbox channel
			room := h.GetRoom(client.RoomID)
			if room == nil {
				continue
			}
			delete(room.Clients, client.ID.String())
			close(client.inbox)
			// TODO: broadcast leave messages to clients

		case message := <-h.broadcast:
			// fanout messages to all connected clients
			room := h.GetRoom(message.RoomID)
			if room == nil {
				continue
			}

			room.Messages = append(room.Messages, message)

			// TODO: persist message in the background

			for _, client := range room.Clients {
				if client.ID == message.SenderID {
					continue
				}

				select {
				case client.inbox <- message:
				// close connection if channel is full
				default:
					close(client.inbox)
					delete(room.Clients, client.ID.String())
				}
			}
		}
	}
}

// GetRoom retrieves a room if present
func (h *Hub) GetRoom(id uuid.UUID) *Room {

	room, ok := h.Rooms[id.String()]
	if !ok {
		return nil
	}
	return room
}
