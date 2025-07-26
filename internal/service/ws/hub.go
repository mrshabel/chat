package ws

import "github.com/mrshabel/chat/internal/model"

// Room holds all connected clients
type Room struct {
	// client id to connection mapping
	Clients  map[string]*Client
	ID       string
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
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[string]*Room),
		broadcast:  make(chan *model.Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
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
				// create room if not present
				room = &Room{
					ID:       client.RoomID,
					Name:     client.RoomID,
					Messages: make([]*model.Message, 0),
					Clients:  make(map[string]*Client),
				}
				h.Rooms[client.RoomID] = room
			}
			room.Clients[client.ID] = client

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
			delete(room.Clients, client.ID)
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
				if client.ID == message.CreatorID {
					continue
				}

				select {
				case client.inbox <- message:
				// close connection if channel is full
				default:
					close(client.inbox)
					delete(room.Clients, client.ID)
				}
			}
		}
	}
}

// GetRoom retrieves a room if present
func (h *Hub) GetRoom(id string) *Room {
	room, ok := h.Rooms[id]
	if !ok {
		return nil
	}
	return room
}
