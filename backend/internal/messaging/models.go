package messaging

import (
	"time"

	"github.com/google/uuid"
)

// Conversation represents a message thread between two agents.
type Conversation struct {
	ID             uuid.UUID  `json:"id"`
	Participant1ID uuid.UUID  `json:"participant_1_id"`
	Participant2ID uuid.UUID  `json:"participant_2_id"`
	// Optional context (what the conversation is about)
	ListingID     *uuid.UUID `json:"listing_id,omitempty"`
	RequestID     *uuid.UUID `json:"request_id,omitempty"`
	AuctionID     *uuid.UUID `json:"auction_id,omitempty"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`

	// Enriched fields (from joins, relative to requesting agent)
	OtherParticipantID   uuid.UUID `json:"other_participant_id,omitempty"`
	OtherParticipantName string    `json:"other_participant_name,omitempty"`
	OtherAvatarURL       *string   `json:"other_avatar_url,omitempty"`
	UnreadCount          int       `json:"unread_count,omitempty"`
	LastMessage          *Message  `json:"last_message,omitempty"`

	// Context info (from joins)
	ListingTitle string `json:"listing_title,omitempty"`
	RequestTitle string `json:"request_title,omitempty"`
	AuctionTitle string `json:"auction_title,omitempty"`
}

// Message represents a single message in a conversation.
type Message struct {
	ID             uuid.UUID  `json:"id"`
	ConversationID uuid.UUID  `json:"conversation_id"`
	SenderID       uuid.UUID  `json:"sender_id"`
	Content        string     `json:"content"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Enriched fields (from agent join)
	SenderName      string  `json:"sender_name,omitempty"`
	SenderAvatarURL *string `json:"sender_avatar_url,omitempty"`
}

// --- Request/Response DTOs ---

// SendMessageRequest is the request body for sending a message.
type SendMessageRequest struct {
	RecipientID uuid.UUID  `json:"recipient_id"`
	Content     string     `json:"content"`
	// Optional context (creates or finds conversation in this context)
	ListingID *uuid.UUID `json:"listing_id,omitempty"`
	RequestID *uuid.UUID `json:"request_id,omitempty"`
	AuctionID *uuid.UUID `json:"auction_id,omitempty"`
}

// ReplyToConversationRequest is the request body for replying to a conversation.
type ReplyToConversationRequest struct {
	Content string `json:"content"`
}

// ListConversationsParams contains filter parameters for listing conversations.
type ListConversationsParams struct {
	Limit  int
	Offset int
}

// ConversationsResponse is the paginated list of conversations.
type ConversationsResponse struct {
	Conversations []Conversation `json:"conversations"`
	Total         int            `json:"total"`
	UnreadTotal   int            `json:"unread_total"`
}

// MessagesResponse is the paginated list of messages.
type MessagesResponse struct {
	Messages []Message `json:"messages"`
	Total    int       `json:"total"`
}
