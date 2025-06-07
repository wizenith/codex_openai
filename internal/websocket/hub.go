package websocket

import (
	"encoding/json"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastTaskUpdate sends a task update to all connected clients
func (h *Hub) BroadcastTaskUpdate(userID int64, updateType string, data interface{}) error {
	message := map[string]interface{}{
		"type":    updateType,
		"user_id": userID,
		"data":    data,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.broadcast <- jsonData
	return nil
}

// BroadcastToUser sends a message to all clients of a specific user
func (h *Hub) BroadcastToUser(userID int64, updateType string, data interface{}) error {
	message := map[string]interface{}{
		"type": updateType,
		"data": data,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.userID == userID {
			select {
			case client.send <- jsonData:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}

	return nil
}

// Register adds a client to the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}