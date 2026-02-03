package messaging

import (
	"context"

	"github.com/google/uuid"
)

// RepositoryInterface defines the database operations for messaging.
type RepositoryInterface interface {
	// Conversations
	CreateConversation(ctx context.Context, conv *Conversation) error
	GetConversationByID(ctx context.Context, id uuid.UUID) (*Conversation, error)
	GetConversationByParticipants(ctx context.Context, agent1ID, agent2ID uuid.UUID, listingID, requestID, auctionID *uuid.UUID) (*Conversation, error)
	GetConversationsByAgentID(ctx context.Context, agentID uuid.UUID, limit, offset int) ([]Conversation, int, error)
	UpdateConversationLastMessage(ctx context.Context, id uuid.UUID) error

	// Messages
	CreateMessage(ctx context.Context, msg *Message) error
	GetMessagesByConversationID(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]Message, int, error)
	MarkMessagesAsRead(ctx context.Context, conversationID, agentID uuid.UUID) error

	// Read status
	GetOrCreateReadStatus(ctx context.Context, conversationID, agentID uuid.UUID) error
	IncrementUnreadCount(ctx context.Context, conversationID, agentID uuid.UUID) error
	ResetUnreadCount(ctx context.Context, conversationID, agentID uuid.UUID) error
	GetTotalUnreadCount(ctx context.Context, agentID uuid.UUID) (int, error)
}

// EventPublisher publishes events to the notification system.
type EventPublisher interface {
	Publish(ctx context.Context, eventType string, payload map[string]any) error
}

// AgentInfo contains minimal agent info needed for messaging.
type AgentInfo struct {
	ID    uuid.UUID
	Name  string
	Email string
}

// AgentGetter retrieves agent information.
type AgentGetter interface {
	GetAgentByID(ctx context.Context, agentID uuid.UUID) (AgentInfo, error)
}
