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

func TestMessageTypes(t *testing.T) {
	tests := []struct {
		msgType string
		payload map[string]any
	}{
		{"offer.received", map[string]any{"offer_id": "123"}},
		{"bid.placed", map[string]any{"bid_id": "456", "amount": 100.0}},
		{"transaction.completed", map[string]any{"transaction_id": "789"}},
		{"request.created", map[string]any{"request_id": "abc"}},
		{"ping", nil},
		{"pong", nil},
	}

	for _, tt := range tests {
		t.Run(tt.msgType, func(t *testing.T) {
			msg := Message{
				Type:    tt.msgType,
				Payload: tt.payload,
			}
			if msg.Type != tt.msgType {
				t.Errorf("expected type %s, got %s", tt.msgType, msg.Type)
			}
		})
	}
}

func TestAgentMessage(t *testing.T) {
	agentID := uuid.New()
	msgData := []byte(`{"type":"test","payload":{}}`)

	agentMsg := &AgentMessage{
		AgentID: agentID,
		Message: msgData,
	}

	if agentMsg.AgentID != agentID {
		t.Error("AgentID not set correctly")
	}
	if string(agentMsg.Message) != string(msgData) {
		t.Error("Message not set correctly")
	}
}

func TestHubConcurrentAccess(t *testing.T) {
	hub := NewHub()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go hub.Run(ctx)

	// Concurrent reads
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				hub.ConnectedCount()
				hub.IsConnected(uuid.New())
			}
			done <- struct{}{}
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestHubMultipleAgentChecks(t *testing.T) {
	hub := NewHub()

	agents := make([]uuid.UUID, 10)
	for i := range agents {
		agents[i] = uuid.New()
	}

	// None should be connected
	for _, agentID := range agents {
		if hub.IsConnected(agentID) {
			t.Errorf("agent %s should not be connected", agentID)
		}
	}

	// Count should be 0
	if hub.ConnectedCount() != 0 {
		t.Errorf("expected 0 connections, got %d", hub.ConnectedCount())
	}
}

func TestMessagePayloadTypes(t *testing.T) {
	msg := Message{
		Type: "complex.payload",
		Payload: map[string]any{
			"string":  "value",
			"int":     42,
			"float":   3.14,
			"bool":    true,
			"nil":     nil,
			"array":   []any{1, 2, 3},
			"nested":  map[string]any{"key": "value"},
		},
	}

	if msg.Payload["string"] != "value" {
		t.Error("string payload incorrect")
	}
	if msg.Payload["int"] != 42 {
		t.Error("int payload incorrect")
	}
	if msg.Payload["float"] != 3.14 {
		t.Error("float payload incorrect")
	}
	if msg.Payload["bool"] != true {
		t.Error("bool payload incorrect")
	}
	if msg.Payload["nil"] != nil {
		t.Error("nil payload incorrect")
	}
}

func TestClientStruct(t *testing.T) {
	hub := NewHub()
	agentID := uuid.New()

	client := &Client{
		hub:     hub,
		conn:    nil, // Would be a real connection in production
		send:    make(chan []byte, 256),
		agentID: agentID,
	}

	if client.hub != hub {
		t.Error("hub not set correctly")
	}
	if client.agentID != agentID {
		t.Error("agentID not set correctly")
	}
	if client.send == nil {
		t.Error("send channel should not be nil")
	}
}

func TestSendToAgentMultipleMessages(t *testing.T) {
	hub := NewHub()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go hub.Run(ctx)

	agentID := uuid.New()

	// Send multiple messages (should not block even if agent not connected)
	for i := 0; i < 10; i++ {
		err := hub.SendToAgent(agentID, Message{
			Type: "test",
			Payload: map[string]any{
				"index": i,
			},
		})
		if err != nil {
			t.Errorf("SendToAgent() error on message %d: %v", i, err)
		}
	}
}

func TestConstants(t *testing.T) {
	// Verify constants are set
	if writeWait <= 0 {
		t.Error("writeWait should be positive")
	}
	if pongWait <= 0 {
		t.Error("pongWait should be positive")
	}
	if pingPeriod <= 0 {
		t.Error("pingPeriod should be positive")
	}
	if maxMessageSize <= 0 {
		t.Error("maxMessageSize should be positive")
	}

	// pingPeriod should be less than pongWait (as per comment in code)
	if pingPeriod >= pongWait {
		t.Error("pingPeriod should be less than pongWait")
	}
}
