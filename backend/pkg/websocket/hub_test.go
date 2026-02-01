package websocket

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()
	if hub == nil {
		t.Fatal("NewHub() returned nil")
	}
	if hub.clients == nil {
		t.Error("clients map is nil")
	}
	if hub.broadcast == nil {
		t.Error("broadcast channel is nil")
	}
	if hub.register == nil {
		t.Error("register channel is nil")
	}
	if hub.unregister == nil {
		t.Error("unregister channel is nil")
	}
}

func TestHubConnectedCount(t *testing.T) {
	hub := NewHub()

	if count := hub.ConnectedCount(); count != 0 {
		t.Errorf("ConnectedCount() = %d, want 0", count)
	}
}

func TestHubIsConnected(t *testing.T) {
	hub := NewHub()
	agentID := uuid.New()

	if hub.IsConnected(agentID) {
		t.Error("IsConnected() should return false for unknown agent")
	}
}

func TestHubRun(t *testing.T) {
	hub := NewHub()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Run hub in background
	done := make(chan struct{})
	go func() {
		hub.Run(ctx)
		close(done)
	}()

	// Wait for context to cancel
	<-ctx.Done()

	// Hub should stop
	select {
	case <-done:
		// Good, hub stopped
	case <-time.After(time.Second):
		t.Error("Hub did not stop after context cancellation")
	}
}

func TestMessageSerialization(t *testing.T) {
	msg := Message{
		Type: "test.event",
		Payload: map[string]any{
			"key": "value",
			"num": 42,
		},
	}

	if msg.Type != "test.event" {
		t.Errorf("Message.Type = %s, want test.event", msg.Type)
	}

	if msg.Payload["key"] != "value" {
		t.Errorf("Message.Payload[key] = %v, want value", msg.Payload["key"])
	}
}

func TestSendToAgentWhenNotConnected(t *testing.T) {
	hub := NewHub()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go hub.Run(ctx)

	// Sending to non-connected agent should not block
	agentID := uuid.New()
	err := hub.SendToAgent(agentID, Message{
		Type:    "test",
		Payload: map[string]any{},
	})

	if err != nil {
		t.Errorf("SendToAgent() error = %v, want nil", err)
	}
}
