package websockets

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
)

// Client represents a single WebSocket connection in the auction system
type Client struct {
	ConnID     string          // Unique connection identifier (exported for setting)
	UserID     string          // ID of the user associated with this connection
	RoomID     string          // Auction room this client is participating in
	connection *websocket.Conn // Underlying WebSocket connection
	manager    *Manager        // Reference to the manager handling this client
	send       chan []byte     // Buffered channel for outbound messages
}

func (c *Client) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.connection.Close()
	}()

	// Configure connection limits
	c.connection.SetReadDeadline(time.Now().Add(pongWait))
	c.connection.SetPongHandler(func(string) error {
		c.connection.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	c.connection.SetReadLimit(maxMessageSize)

	for {
		// Read incoming message from the WebSocket
		_, message, err := c.connection.ReadMessage()
		if err != nil {
			// Connection error or closed - exit the loop
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for client %s: %v", c.ConnID, err)
			}
			break
		}
		// Forward message to manager for broadcasting to clients in the same room
		c.manager.broadcastToRoom(c.RoomID, message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.connection.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.connection.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Channel closed - send close message and exit
				c.connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Write message to the WebSocket connection
			err := c.connection.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("Write error for client %s: %v", c.ConnID, err)
				return
			}

		case <-ticker.C:
			// Send periodic ping to keep connection alive
			c.connection.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
