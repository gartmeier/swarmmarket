package messaging

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// EmailQueuer queues emails for async delivery.
type EmailQueuer interface {
	QueueEmail(ctx context.Context, recipientEmail string, recipientAgentID uuid.UUID, template string, payload map[string]any) error
}

// Service handles messaging business logic.
type Service struct {
	repo        RepositoryInterface
	publisher   EventPublisher
	emailQueuer EmailQueuer
	agentGetter AgentGetter
}

// NewService creates a new messaging service.
func NewService(repo RepositoryInterface, publisher EventPublisher, emailQueuer EmailQueuer, agentGetter AgentGetter) *Service {
	return &Service{
		repo:        repo,
		publisher:   publisher,
		emailQueuer: emailQueuer,
		agentGetter: agentGetter,
	}
}

// SendMessage sends a message to another agent, creating a conversation if needed.
func (s *Service) SendMessage(ctx context.Context, senderID uuid.UUID, req *SendMessageRequest) (*Message, *Conversation, error) {
	// Validate input
	if req.Content == "" {
		return nil, nil, fmt.Errorf("message content is required")
	}
	if len(req.Content) > 5000 {
		return nil, nil, fmt.Errorf("message content too long (max 5000 characters)")
	}
	if req.RecipientID == senderID {
		return nil, nil, fmt.Errorf("cannot send message to yourself")
	}

	// Find or create conversation
	conv, err := s.repo.GetConversationByParticipants(ctx, senderID, req.RecipientID, req.ListingID, req.RequestID, req.AuctionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check for existing conversation: %w", err)
	}

	now := time.Now().UTC()

	if conv == nil {
		// Create new conversation
		conv = &Conversation{
			ID:             uuid.New(),
			Participant1ID: senderID,
			Participant2ID: req.RecipientID,
			ListingID:      req.ListingID,
			RequestID:      req.RequestID,
			AuctionID:      req.AuctionID,
			CreatedAt:      now,
		}
		if err := s.repo.CreateConversation(ctx, conv); err != nil {
			return nil, nil, fmt.Errorf("failed to create conversation: %w", err)
		}

		// Initialize read status for both participants
		_ = s.repo.GetOrCreateReadStatus(ctx, conv.ID, senderID)
		_ = s.repo.GetOrCreateReadStatus(ctx, conv.ID, req.RecipientID)
	}

	// Create message
	msg := &Message{
		ID:             uuid.New(),
		ConversationID: conv.ID,
		SenderID:       senderID,
		Content:        req.Content,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return nil, nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Update conversation last_message_at
	_ = s.repo.UpdateConversationLastMessage(ctx, conv.ID)

	// Increment unread count for recipient
	_ = s.repo.IncrementUnreadCount(ctx, conv.ID, req.RecipientID)

	// Publish event for real-time notification + email
	s.publishEvent(ctx, "message.received", map[string]any{
		"message_id":       msg.ID,
		"conversation_id":  conv.ID,
		"sender_id":        senderID,
		"recipient_id":     req.RecipientID,
		"content_preview":  truncate(req.Content, 100),
		"listing_id":       req.ListingID,
		"request_id":       req.RequestID,
		"auction_id":       req.AuctionID,
	})

	// Queue email notification for recipient
	s.queueMessageEmail(ctx, senderID, req.RecipientID, conv.ID, req.Content)

	return msg, conv, nil
}

// ReplyToConversation adds a message to an existing conversation.
func (s *Service) ReplyToConversation(ctx context.Context, senderID, conversationID uuid.UUID, req *ReplyToConversationRequest) (*Message, error) {
	// Validate input
	if req.Content == "" {
		return nil, fmt.Errorf("message content is required")
	}
	if len(req.Content) > 5000 {
		return nil, fmt.Errorf("message content too long (max 5000 characters)")
	}

	// Get conversation and verify sender is a participant
	conv, err := s.repo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	if conv.Participant1ID != senderID && conv.Participant2ID != senderID {
		return nil, fmt.Errorf("not authorized to reply to this conversation")
	}

	// Determine recipient
	recipientID := conv.Participant1ID
	if conv.Participant1ID == senderID {
		recipientID = conv.Participant2ID
	}

	now := time.Now().UTC()

	// Create message
	msg := &Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        req.Content,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Update conversation last_message_at
	_ = s.repo.UpdateConversationLastMessage(ctx, conversationID)

	// Increment unread count for recipient
	_ = s.repo.IncrementUnreadCount(ctx, conversationID, recipientID)

	// Publish event
	s.publishEvent(ctx, "message.received", map[string]any{
		"message_id":       msg.ID,
		"conversation_id":  conversationID,
		"sender_id":        senderID,
		"recipient_id":     recipientID,
		"content_preview":  truncate(req.Content, 100),
		"listing_id":       conv.ListingID,
		"request_id":       conv.RequestID,
		"auction_id":       conv.AuctionID,
	})

	// Queue email notification for recipient
	s.queueMessageEmail(ctx, senderID, recipientID, conversationID, req.Content)

	return msg, nil
}

// GetConversations returns all conversations for an agent.
func (s *Service) GetConversations(ctx context.Context, agentID uuid.UUID, params ListConversationsParams) (*ConversationsResponse, error) {
	conversations, total, err := s.repo.GetConversationsByAgentID(ctx, agentID, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}

	// Calculate total unread
	unreadTotal, _ := s.repo.GetTotalUnreadCount(ctx, agentID)

	return &ConversationsResponse{
		Conversations: conversations,
		Total:         total,
		UnreadTotal:   unreadTotal,
	}, nil
}

// GetConversation returns a single conversation with authorization check.
func (s *Service) GetConversation(ctx context.Context, agentID, conversationID uuid.UUID) (*Conversation, error) {
	conv, err := s.repo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	if conv.Participant1ID != agentID && conv.Participant2ID != agentID {
		return nil, fmt.Errorf("not authorized to view this conversation")
	}

	return conv, nil
}

// GetMessages returns messages in a conversation.
func (s *Service) GetMessages(ctx context.Context, agentID, conversationID uuid.UUID, limit, offset int) (*MessagesResponse, error) {
	// Verify access
	conv, err := s.repo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if conv.Participant1ID != agentID && conv.Participant2ID != agentID {
		return nil, fmt.Errorf("not authorized to view this conversation")
	}

	messages, total, err := s.repo.GetMessagesByConversationID(ctx, conversationID, limit, offset)
	if err != nil {
		return nil, err
	}

	return &MessagesResponse{
		Messages: messages,
		Total:    total,
	}, nil
}

// MarkAsRead marks all messages in a conversation as read for the agent.
func (s *Service) MarkAsRead(ctx context.Context, agentID, conversationID uuid.UUID) error {
	// Verify access
	conv, err := s.repo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return err
	}
	if conv.Participant1ID != agentID && conv.Participant2ID != agentID {
		return fmt.Errorf("not authorized")
	}

	// Mark messages as read
	if err := s.repo.MarkMessagesAsRead(ctx, conversationID, agentID); err != nil {
		return err
	}

	// Reset unread count
	if err := s.repo.ResetUnreadCount(ctx, conversationID, agentID); err != nil {
		return err
	}

	// Publish read event
	otherID := conv.Participant1ID
	if conv.Participant1ID == agentID {
		otherID = conv.Participant2ID
	}
	s.publishEvent(ctx, "message.read", map[string]any{
		"conversation_id": conversationID,
		"reader_id":       agentID,
		"other_id":        otherID,
	})

	return nil
}

// GetUnreadCount returns the total unread message count for an agent.
func (s *Service) GetUnreadCount(ctx context.Context, agentID uuid.UUID) (int, error) {
	return s.repo.GetTotalUnreadCount(ctx, agentID)
}

// publishEvent publishes an event asynchronously.
func (s *Service) publishEvent(ctx context.Context, eventType string, payload map[string]any) {
	if s.publisher != nil {
		go s.publisher.Publish(context.Background(), eventType, payload)
	}
}

// truncate truncates a string to max length with ellipsis.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// queueMessageEmail queues an email notification for the message recipient.
func (s *Service) queueMessageEmail(ctx context.Context, senderID, recipientID, conversationID uuid.UUID, content string) {
	if s.emailQueuer == nil || s.agentGetter == nil {
		return
	}

	// Get sender and recipient info
	sender, err := s.agentGetter.GetAgentByID(ctx, senderID)
	if err != nil {
		log.Printf("[MESSAGING] Failed to get sender for email: %v", err)
		return
	}

	recipient, err := s.agentGetter.GetAgentByID(ctx, recipientID)
	if err != nil {
		log.Printf("[MESSAGING] Failed to get recipient for email: %v", err)
		return
	}

	// Only queue email if recipient has an email address
	if recipient.Email == "" {
		return
	}

	// Queue the email (template is "new_message")
	err = s.emailQueuer.QueueEmail(ctx, recipient.Email, recipientID, "new_message", map[string]any{
		"recipient_name":  recipient.Name,
		"sender_name":     sender.Name,
		"content_preview": truncate(content, 200),
		"conversation_id": conversationID.String(),
	})
	if err != nil {
		log.Printf("[MESSAGING] Failed to queue email: %v", err)
	}
}
