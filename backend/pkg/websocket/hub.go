package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Message represents a WebSocket message.
type Message struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

// Client represents a connected WebSocket client.
type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	send    chan []byte
	agentID uuid.UUID
}

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	// Registered clients by agent ID
	clients map[uuid.UUID]*Client

	// Channel for broadcasting to specific agents
	broadcast chan *AgentMessage

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe client access
	mu sync.RWMutex
}

// AgentMessage is a message targeted at a specific agent.
type AgentMessage struct {
	AgentID uuid.UUID
	Message []byte
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]*Client),
		broadcast:  make(chan *AgentMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			// Close existing connection if any
			if existing, ok := h.clients[client.agentID]; ok {
				close(existing.send)
				existing.conn.Close()
			}
			h.clients[client.agentID] = client
			h.mu.Unlock()
			log.Printf("WebSocket: Agent %s connected", client.agentID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.agentID]; ok {
				delete(h.clients, client.agentID)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("WebSocket: Agent %s disconnected", client.agentID)

		case msg := <-h.broadcast:
			h.mu.RLock()
			if client, ok := h.clients[msg.AgentID]; ok {
				select {
				case client.send <- msg.Message:
				default:
					// Client buffer full, skip message
					log.Printf("WebSocket: Buffer full for agent %s, dropping message", msg.AgentID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// SendToAgent sends a message to a specific agent.
func (h *Hub) SendToAgent(agentID uuid.UUID, message Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.broadcast <- &AgentMessage{
		AgentID: agentID,
		Message: data,
	}
	return nil
}

// IsConnected checks if an agent is currently connected.
func (h *Hub) IsConnected(agentID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[agentID]
	return ok
}

// ConnectedCount returns the number of connected clients.
func (h *Hub) ConnectedCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages (e.g., subscribe to specific events)
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// Process message based on type
		switch msg.Type {
		case "ping":
			c.send <- []byte(`{"type":"pong"}`)
		case "subscribe":
			// Could implement topic subscriptions here
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
