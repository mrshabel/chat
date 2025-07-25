package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeTimeout = 10 * time.Second
	// time allowed to read next pong message from peer
	pongTimeout = 1 * time.Minute

	// interval to ping clients
	pingInterval = 30 * time.Second

	// maximum message size
	maxMessageSize = 512
)

// websocket connection upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// allow all client origins
		return true
	},
	EnableCompression: true,
}

// Client holds the websocket connection and while connecting it  with the hub
type Client struct {
	// communication channel with central hub
	hub  *Hub
	conn *websocket.Conn
	// channel to receive messages
	inbox chan *Message

	// currently joined room
	RoomID string
	// client information
	ID       string
	Username string
}

// readPump sends message from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		// unregister the client and close the websocket connection
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// set message and timeout defaults
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongTimeout))
	// pong handler (client heartbeat response)
	c.conn.SetPongHandler(func(appData string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})

	// read messages from client
	for {
		_, m, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// clean message and broadcast it
		message := &Message{
			Content:         string(m),
			RoomID:          c.RoomID,
			CreatorID:       c.ID,
			CreatorUsername: c.Username,
			CreatedAt:       time.Now().UTC(),
		}
		c.hub.broadcast <- message
	}
}

// writePump sends messages from the hub to the current websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		// message received on client's channel
		case message, ok := <-c.inbox:
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			// channel closed by hub so we close the connection
			if !ok {
				// c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.WriteJSON(message)
		case <-ticker.C:
			// ping client
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWS handles the websocket requests from peer
func serveWS(hub *Hub, c *Client, w http.ResponseWriter, r *http.Request) {
	// upgrade client http connection to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade connection %v\n", err)
		return
	}

	client := &Client{hub: hub, conn: conn, inbox: make(chan *Message), RoomID: c.RoomID, ID: c.ID, Username: c.Username}
	client.hub.register <- client

	// handle connection reads and writes
	go client.writePump()
	go client.readPump()
}
