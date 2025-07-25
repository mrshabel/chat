package main

// Hub holds the set of active clients and broadcasts messages to them
type Hub struct {
	clients map[*Client]struct{}

	// inbound messages from clients
	broadcast chan []byte

	// register/enter requests from client
	register chan *Client

	// unregister/leave request from client
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// run starts the hub and handles all related events
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = struct{}{}
			// broadcast message to all other clients in hub
		case client := <-h.unregister:
			delete(h.clients, client)
			close(client.send)
			// broadcast leave messages to clients
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				// close connection if channel is full
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
