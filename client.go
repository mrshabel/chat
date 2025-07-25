package main

import (
	"bytes"
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

// message delimiters
var (
	newline = []byte{'\n'}
	space   = []byte{' '}
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
	hub  *Hub
	conn *websocket.Conn

	// outbound message channel
	send chan []byte
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
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// clean message and broadcast it
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
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
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			// channel closed by hub so we close the connection
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			writer, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			writer.Write(message)
			// add queued message to the connect websocket message
			n := len(c.send)
			for range n {
				writer.Write(newline)
				writer.Write(<-c.send)
			}

			if err := writer.Close(); err != nil {
				return
			}

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
func serveWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// upgrade client http connection to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade connection %v\n", err)
		return
	}

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// handle connection reads and writes
	go client.writePump()
	go client.readPump()
}
