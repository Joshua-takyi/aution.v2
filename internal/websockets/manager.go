package websockets

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: In production, validate against allowed origins
		// origin := r.Header.Get("Origin")
		// return origin == "https://yourdomain.com"
		return true
	},
}

type Message struct {
	Data   []byte
	RoomID string
}

type Ticket struct {
	UserID    string
	ExpiresAt time.Time
}

type Manager struct {
	clients    map[*Client]bool
	register   chan *Client
	mu         sync.RWMutex
	unregister chan *Client
	broadcast  chan Message
	tickets    map[string]Ticket
}

func NewManager() *Manager {
	return &Manager{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		broadcast:  make(chan Message),
		unregister: make(chan *Client),
		tickets:    make(map[string]Ticket),
	}
}

func (m *Manager) CreateTicket(userID string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	ticket := uuid.New().String()
	m.tickets[ticket] = Ticket{
		UserID:    userID,
		ExpiresAt: time.Now().Add(30 * time.Second),
	}
	return ticket
}

func (m *Manager) ConsumeTicket(ticketID string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ticket, ok := m.tickets[ticketID]
	if !ok {
		return "", false
	}

	delete(m.tickets, ticketID)

	if time.Now().After(ticket.ExpiresAt) {
		return "", false
	}

	return ticket.UserID, true
}

func (manager *Manager) Start() {
	go func() {
		for {
			select {
			case client := <-manager.register:
				manager.mu.Lock()
				manager.clients[client] = true
				manager.mu.Unlock()
			case client := <-manager.unregister:
				manager.mu.Lock()

				if _, ok := manager.clients[client]; ok {
					delete(manager.clients, client)
					close(client.send)
				}
				fmt.Printf("[WEBSOCKET] unregistered user %s\n", client.UserID)
				manager.mu.Unlock()

			case msg := <-manager.broadcast:
				manager.mu.RLock()
				for client := range manager.clients {
					if msg.RoomID == "" || client.RoomID == msg.RoomID {
						select {
						case client.send <- msg.Data:
							// Message queued successfully
						default:
							// Client's send channel is full - remove it
							go func(c *Client) {
								manager.unregister <- c
							}(client)
						}

					}
				}
				manager.mu.RUnlock()
			}

		}
	}()
}

func (m *Manager) Run(ctx *gin.Context) {
	ticketID := ctx.Query("ticket")
	if ticketID == "" {
		log.Printf("WebSocket connection missing ticket")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Ticket required"})
		return
	}

	userID, ok := m.ConsumeTicket(ticketID)
	if !ok {
		log.Printf("Invalid or expired WebSocket ticket: %s", ticketID)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired ticket"})
		return
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	roomID := ctx.Param("id")
	if roomID == "" {
		log.Printf("WebSocket connection missing auction ID parameter")
		conn.Close()
		return
	}

	// Create new client for this connection
	client := &Client{
		ConnID:     uuid.New().String(),
		UserID:     userID,
		RoomID:     roomID,
		connection: conn,
		manager:    m,
		send:       make(chan []byte, 256), // Buffered to prevent blocking
	}

	// Register the client with the manager
	m.register <- client

	go client.writePump()
	go client.readPump()
}

func (manager *Manager) Broadcast(message []byte) {
	manager.broadcast <- Message{RoomID: "", Data: message}
}

func (manager *Manager) broadcastToRoom(roomID string, message []byte) {
	manager.broadcast <- Message{RoomID: roomID, Data: message}
}

func (manager *Manager) GetClientCount() int {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return len(manager.clients)
}

func (manager *Manager) GetRoomClientCount(roomID string) int {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	count := 0
	for client := range manager.clients {
		if client.RoomID == roomID {
			count++
		}
	}
	return count
}

func (manager *Manager) SendNotificationToUser(userID string, notification Notification) {
	data, err := notification.ToMessage()
	if err != nil {
		log.Printf("[WEBSOCKET] failed to marshal notification: %v", err)
		return
	}

	manager.mu.RLock()
	defer manager.mu.RUnlock()
	for client := range manager.clients {
		if client.UserID == userID {
			select {
			case client.send <- data:
			default:
				go func(c *Client) {
					manager.unregister <- c
				}(client)
			}
		}
	}
}

func (manager *Manager) BroadcastNotificationToRoom(roomID string, notification Notification) {
	data, err := notification.ToMessage()
	if err != nil {
		log.Printf("[WEBSOCKET] failed to marshal notification: %v", err)
		return
	}

	manager.broadcast <- Message{RoomID: roomID, Data: data}
}

// Stop gracefully shuts down the manager by closing all channels
func (manager *Manager) Stop() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// Close all client connections
	for client := range manager.clients {
		close(client.send)
		client.connection.Close()
	}

	// Close manager channels
	close(manager.register)
	close(manager.unregister)
	close(manager.broadcast)
}
