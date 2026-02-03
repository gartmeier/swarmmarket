package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles database operations for messaging.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new messaging repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// CreateConversation creates a new conversation.
func (r *Repository) CreateConversation(ctx context.Context, conv *Conversation) error {
	query := `
		INSERT INTO conversations (id, participant_1_id, participant_2_id, listing_id, request_id, auction_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query,
		conv.ID, conv.Participant1ID, conv.Participant2ID,
		conv.ListingID, conv.RequestID, conv.AuctionID, conv.CreatedAt,
	)
	return err
}

// GetConversationByID retrieves a conversation by ID.
func (r *Repository) GetConversationByID(ctx context.Context, id uuid.UUID) (*Conversation, error) {
	query := `
		SELECT id, participant_1_id, participant_2_id, listing_id, request_id, auction_id, last_message_at, created_at
		FROM conversations
		WHERE id = $1
	`
	var conv Conversation
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&conv.ID, &conv.Participant1ID, &conv.Participant2ID,
		&conv.ListingID, &conv.RequestID, &conv.AuctionID,
		&conv.LastMessageAt, &conv.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("conversation not found")
	}
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// GetConversationByParticipants finds an existing conversation between two agents with optional context.
func (r *Repository) GetConversationByParticipants(ctx context.Context, agent1ID, agent2ID uuid.UUID, listingID, requestID, auctionID *uuid.UUID) (*Conversation, error) {
	// Normalize the lookup - participants can be in either order
	query := `
		SELECT id, participant_1_id, participant_2_id, listing_id, request_id, auction_id, last_message_at, created_at
		FROM conversations
		WHERE (
			(participant_1_id = $1 AND participant_2_id = $2) OR
			(participant_1_id = $2 AND participant_2_id = $1)
		)
		AND (listing_id IS NOT DISTINCT FROM $3)
		AND (request_id IS NOT DISTINCT FROM $4)
		AND (auction_id IS NOT DISTINCT FROM $5)
	`
	var conv Conversation
	err := r.pool.QueryRow(ctx, query, agent1ID, agent2ID, listingID, requestID, auctionID).Scan(
		&conv.ID, &conv.Participant1ID, &conv.Participant2ID,
		&conv.ListingID, &conv.RequestID, &conv.AuctionID,
		&conv.LastMessageAt, &conv.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil // No conversation found
	}
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// GetConversationsByAgentID retrieves all conversations for an agent with enriched info.
func (r *Repository) GetConversationsByAgentID(ctx context.Context, agentID uuid.UUID, limit, offset int) ([]Conversation, int, error) {
	if limit <= 0 {
		limit = 20
	}

	// Count total
	countQuery := `
		SELECT COUNT(*) FROM conversations
		WHERE participant_1_id = $1 OR participant_2_id = $1
	`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, agentID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get conversations with enriched data
	query := `
		SELECT
			c.id, c.participant_1_id, c.participant_2_id,
			c.listing_id, c.request_id, c.auction_id,
			c.last_message_at, c.created_at,
			-- Other participant info
			CASE WHEN c.participant_1_id = $1 THEN c.participant_2_id ELSE c.participant_1_id END as other_id,
			COALESCE(a.name, '') as other_name,
			a.avatar_url as other_avatar,
			-- Unread count
			COALESCE(crs.unread_count, 0) as unread_count,
			-- Context titles
			COALESCE(l.title, '') as listing_title,
			COALESCE(r.title, '') as request_title,
			COALESCE(auc.title, '') as auction_title,
			-- Last message
			m.id as last_msg_id, m.sender_id as last_msg_sender, m.content as last_msg_content, m.created_at as last_msg_at
		FROM conversations c
		LEFT JOIN agents a ON a.id = CASE WHEN c.participant_1_id = $1 THEN c.participant_2_id ELSE c.participant_1_id END
		LEFT JOIN conversation_read_status crs ON crs.conversation_id = c.id AND crs.agent_id = $1
		LEFT JOIN listings l ON l.id = c.listing_id
		LEFT JOIN requests r ON r.id = c.request_id
		LEFT JOIN auctions auc ON auc.id = c.auction_id
		LEFT JOIN LATERAL (
			SELECT id, sender_id, content, created_at
			FROM messages
			WHERE conversation_id = c.id
			ORDER BY created_at DESC
			LIMIT 1
		) m ON true
		WHERE c.participant_1_id = $1 OR c.participant_2_id = $1
		ORDER BY COALESCE(c.last_message_at, c.created_at) DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, agentID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var conv Conversation
		var lastMsgID *uuid.UUID
		var lastMsgSender *uuid.UUID
		var lastMsgContent *string
		var lastMsgAt *time.Time

		if err := rows.Scan(
			&conv.ID, &conv.Participant1ID, &conv.Participant2ID,
			&conv.ListingID, &conv.RequestID, &conv.AuctionID,
			&conv.LastMessageAt, &conv.CreatedAt,
			&conv.OtherParticipantID, &conv.OtherParticipantName, &conv.OtherAvatarURL,
			&conv.UnreadCount,
			&conv.ListingTitle, &conv.RequestTitle, &conv.AuctionTitle,
			&lastMsgID, &lastMsgSender, &lastMsgContent, &lastMsgAt,
		); err != nil {
			return nil, 0, err
		}

		// Build last message if exists
		if lastMsgID != nil {
			conv.LastMessage = &Message{
				ID:             *lastMsgID,
				ConversationID: conv.ID,
				SenderID:       *lastMsgSender,
				Content:        *lastMsgContent,
				CreatedAt:      *lastMsgAt,
			}
		}

		conversations = append(conversations, conv)
	}

	return conversations, total, nil
}

// UpdateConversationLastMessage updates the last_message_at timestamp.
func (r *Repository) UpdateConversationLastMessage(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE conversations SET last_message_at = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), id)
	return err
}

// CreateMessage creates a new message.
func (r *Repository) CreateMessage(ctx context.Context, msg *Message) error {
	query := `
		INSERT INTO messages (id, conversation_id, sender_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query,
		msg.ID, msg.ConversationID, msg.SenderID, msg.Content,
		msg.CreatedAt, msg.UpdatedAt,
	)
	return err
}

// GetMessagesByConversationID retrieves messages in a conversation.
func (r *Repository) GetMessagesByConversationID(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]Message, int, error) {
	if limit <= 0 {
		limit = 50
	}

	// Count total
	countQuery := `SELECT COUNT(*) FROM messages WHERE conversation_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, conversationID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get messages with sender info
	query := `
		SELECT m.id, m.conversation_id, m.sender_id, m.content, m.read_at, m.created_at, m.updated_at,
			COALESCE(a.name, '') as sender_name,
			a.avatar_url as sender_avatar
		FROM messages m
		LEFT JOIN agents a ON a.id = m.sender_id
		WHERE m.conversation_id = $1
		ORDER BY m.created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, conversationID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Content,
			&msg.ReadAt, &msg.CreatedAt, &msg.UpdatedAt,
			&msg.SenderName, &msg.SenderAvatarURL,
		); err != nil {
			return nil, 0, err
		}
		messages = append(messages, msg)
	}

	return messages, total, nil
}

// MarkMessagesAsRead marks all messages as read for a recipient.
func (r *Repository) MarkMessagesAsRead(ctx context.Context, conversationID, agentID uuid.UUID) error {
	query := `
		UPDATE messages
		SET read_at = $1
		WHERE conversation_id = $2 AND sender_id != $3 AND read_at IS NULL
	`
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), conversationID, agentID)
	return err
}

// GetOrCreateReadStatus ensures a read status record exists.
func (r *Repository) GetOrCreateReadStatus(ctx context.Context, conversationID, agentID uuid.UUID) error {
	query := `
		INSERT INTO conversation_read_status (id, conversation_id, agent_id, unread_count)
		VALUES ($1, $2, $3, 0)
		ON CONFLICT (conversation_id, agent_id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, uuid.New(), conversationID, agentID)
	return err
}

// IncrementUnreadCount increments the unread count for an agent in a conversation.
func (r *Repository) IncrementUnreadCount(ctx context.Context, conversationID, agentID uuid.UUID) error {
	query := `
		INSERT INTO conversation_read_status (id, conversation_id, agent_id, unread_count)
		VALUES ($1, $2, $3, 1)
		ON CONFLICT (conversation_id, agent_id)
		DO UPDATE SET unread_count = conversation_read_status.unread_count + 1
	`
	_, err := r.pool.Exec(ctx, query, uuid.New(), conversationID, agentID)
	return err
}

// ResetUnreadCount resets the unread count and updates last_read_at.
func (r *Repository) ResetUnreadCount(ctx context.Context, conversationID, agentID uuid.UUID) error {
	query := `
		UPDATE conversation_read_status
		SET unread_count = 0, last_read_at = $1
		WHERE conversation_id = $2 AND agent_id = $3
	`
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), conversationID, agentID)
	return err
}

// GetTotalUnreadCount returns the total unread message count for an agent.
func (r *Repository) GetTotalUnreadCount(ctx context.Context, agentID uuid.UUID) (int, error) {
	query := `
		SELECT COALESCE(SUM(unread_count), 0)
		FROM conversation_read_status
		WHERE agent_id = $1
	`
	var count int
	err := r.pool.QueryRow(ctx, query, agentID).Scan(&count)
	return count, err
}

// Verify interface compliance
var _ RepositoryInterface = (*Repository)(nil)
