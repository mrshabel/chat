package ws

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/mrshabel/chat/internal/model"
	"github.com/mrshabel/chat/internal/service"
)

const (
	MaxMessageLimit = 20
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
	Broadcast chan *model.Message

	// register/enter requests from client
	Register chan *Client

	// unregister/leave request from client
	Unregister chan *Client

	roomService    *service.RoomService
	messageService *service.MessageService
}

func NewHub(roomService *service.RoomService, messageService *service.MessageService) *Hub {
	return &Hub{
		Rooms:          make(map[string]*Room),
		Broadcast:      make(chan *model.Message),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		roomService:    roomService,
		messageService: messageService,
	}
}

// run starts the hub and handles all related events
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// joined specified room and inform members
			room := h.GetRoom(client.RoomID)
			if room == nil {
				continue
			}
			room.Clients[client.ID.String()] = client

			// load recent messages from db and replay to client
			messages, err := h.messageService.GetByRoomID(context.Background(), room.ID, MaxMessageLimit, 0)
			if err != nil {
				log.Printf("failed to load recent messages for room (%s) from db\n", client.RoomID)
			} else {
				room.Messages = messages
			}

			// replay messages history to client
			go func() {
				for _, message := range room.Messages {
					client.Inbox <- message
				}
			}()

		case client := <-h.Unregister:
			// remove client from room and close inbox channel
			room := h.GetRoom(client.RoomID)
			if room == nil {
				continue
			}
			delete(room.Clients, client.ID.String())
			close(client.Inbox)

		case message := <-h.Broadcast:
			// fanout messages to all connected clients
			room := h.GetRoom(message.RoomID)
			if room == nil {
				continue
			}

			// persist message in the background and update inmem messages
			go func() {
				message, err := h.messageService.Create(context.Background(), message)
				if err != nil {
					log.Printf("failed to persist client message: %v\n", err)
					return
				}
				room.Messages = append(room.Messages, message)
			}()

			for _, client := range room.Clients {
				if client.ID == message.SenderID {
					continue
				}
				select {
				case client.Inbox <- message:
				// close connection if inbox channel is full
				default:
					close(client.Inbox)
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
