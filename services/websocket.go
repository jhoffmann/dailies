package services

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketEventType represents the type of WebSocket event
type WebSocketEventType string

const (
	EventTaskReset  WebSocketEventType = "task_reset"
	EventTaskUpdate WebSocketEventType = "task_update"
	EventTaskCreate WebSocketEventType = "task_create"
	EventTaskDelete WebSocketEventType = "task_delete"
	EventTagUpdate  WebSocketEventType = "tag_update"
	EventTagCreate  WebSocketEventType = "tag_create"
	EventTagDelete  WebSocketEventType = "tag_delete"
	EventFreqUpdate WebSocketEventType = "frequency_update"
	EventFreqCreate WebSocketEventType = "frequency_create"
	EventFreqDelete WebSocketEventType = "frequency_delete"
)

// WebSocketEvent represents a WebSocket event
type WebSocketEvent struct {
	Type WebSocketEventType `json:"type"`
	Data any                `json:"data"`
}

// WebSocketManager manages WebSocket connections and broadcasting
type WebSocketManager struct {
	clients    map[*websocket.Conn]bool
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	broadcast  chan WebSocketEvent
	mutex      sync.RWMutex
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		clients:    make(map[*websocket.Conn]bool),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		broadcast:  make(chan WebSocketEvent),
	}
}

// Run starts the WebSocket manager
func (manager *WebSocketManager) Run() {
	for {
		select {
		case client := <-manager.register:
			manager.mutex.Lock()
			manager.clients[client] = true
			manager.mutex.Unlock()
			log.Printf("WebSocket client connected. Total clients: %d", len(manager.clients))

		case client := <-manager.unregister:
			manager.mutex.Lock()
			if _, ok := manager.clients[client]; ok {
				delete(manager.clients, client)
				client.Close()
			}
			manager.mutex.Unlock()
			log.Printf("WebSocket client disconnected. Total clients: %d", len(manager.clients))

		case event := <-manager.broadcast:
			manager.mutex.RLock()
			for client := range manager.clients {
				err := client.WriteJSON(event)
				if err != nil {
					log.Printf("WebSocket write error: %v", err)
					client.Close()
					delete(manager.clients, client)
				}
			}
			manager.mutex.RUnlock()
		}
	}
}

// Broadcast sends an event to all connected clients
func (manager *WebSocketManager) Broadcast(eventType WebSocketEventType, data any) {
	event := WebSocketEvent{
		Type: eventType,
		Data: data,
	}

	select {
	case manager.broadcast <- event:
	default:
		log.Println("WebSocket broadcast channel full, dropping message")
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin (adjust for production)
		return true
	},
}

// HandleWebSocket handles WebSocket connections
func (manager *WebSocketManager) HandleWebSocket() gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}

		manager.register <- conn

		// Handle incoming messages (ping/pong, etc.)
		go func() {
			defer func() {
				manager.unregister <- conn
			}()

			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("WebSocket read error: %v", err)
					}
					break
				}
			}
		}()
	}
}
