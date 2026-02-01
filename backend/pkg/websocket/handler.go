package websocket

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now - in production, restrict this
		return true
	},
}

// Handler handles WebSocket connections.
type Handler struct {
	hub          *Hub
	authenticate func(r *http.Request) (uuid.UUID, error)
}

// NewHandler creates a new WebSocket handler.
func NewHandler(hub *Hub, authenticate func(r *http.Request) (uuid.UUID, error)) *Handler {
	return &Handler{
		hub:          hub,
		authenticate: authenticate,
	}
}

// ServeHTTP handles WebSocket upgrade requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Authenticate the agent
	agentID, err := h.authenticate(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Create client
	client := &Client{
		hub:     h.hub,
		conn:    conn,
		send:    make(chan []byte, 256),
		agentID: agentID,
	}

	// Register client
	h.hub.register <- client

	// Start client goroutines
	go client.writePump()
	go client.readPump()
}
