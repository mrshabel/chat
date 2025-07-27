package ws

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/mrshabel/chat/internal/model"
	"github.com/mrshabel/chat/internal/util"
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

// websocket connection
var Upgrader = websocket.Upgrader{
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
	Hub  *Hub
	Conn *websocket.Conn
	// channel to receive messages
	Inbox chan *model.Message

	// currently joined room
	RoomID uuid.UUID
	// client information
	ID       uuid.UUID
	Username string
}

// ReadPump sends message from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		// unregister the client and close the websocket connection
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	// set message and timeout defaults
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongTimeout))
	// pong handler (client heartbeat response)
	c.Conn.SetPongHandler(func(appData string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})

	// read messages from client
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// clean message and broadcast it
		message := &model.Message{
			Content:        util.SanitizeWSMessage(msg),
			RoomID:         c.RoomID,
			SenderID:       c.ID,
			SenderUsername: c.Username,
		}
		c.Hub.Broadcast <- message
	}
}

// WritePump sends messages from the hub to the current websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		// message received on client's channel
		case message, ok := <-c.Inbox:
			c.Conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			// channel closed by hub so we close the connection
			if !ok {
				// c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.WriteJSON(message)
		case <-ticker.C:
			// ping client
			c.Conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWS handles the websocket requests from peer
func ServeWS(hub *Hub, c *Client, w http.ResponseWriter, r *http.Request) {
	// upgrade client http connection to websocket
	Conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade connection %v\n", err)
		return
	}

	client := &Client{
		Hub:      hub,
		Conn:     Conn,
		Inbox:    make(chan *model.Message),
		RoomID:   c.RoomID,
		ID:       c.ID,
		Username: c.Username,
	}
	client.Hub.Register <- client

	// handle connection reads and writes
	go client.WritePump()
	go client.ReadPump()
}
