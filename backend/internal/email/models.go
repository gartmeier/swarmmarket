package email

import (
	"time"

	"github.com/google/uuid"
)

// EmailTemplate defines which template to use for an email.
type EmailTemplate string

const (
	// Agent events
	TemplateAgentRegistered EmailTemplate = "agent_registered"
	TemplateAgentClaimed    EmailTemplate = "agent_claimed"

	// Message events
	TemplateNewMessage EmailTemplate = "new_message"

	// Marketplace events
	TemplateListingCreated   EmailTemplate = "listing_created"
	TemplateListingPurchased EmailTemplate = "listing_purchased"
	TemplateNewComment       EmailTemplate = "new_comment"

	// Request/Offer events
	TemplateRequestCreated EmailTemplate = "request_created"
	TemplateOfferReceived  EmailTemplate = "offer_received"
	TemplateOfferAccepted  EmailTemplate = "offer_accepted"
	TemplateOfferRejected  EmailTemplate = "offer_rejected"

	// Auction events
	TemplateAuctionCreated    EmailTemplate = "auction_created"
	TemplateNewBid            EmailTemplate = "new_bid"
	TemplateOutbid            EmailTemplate = "outbid"
	TemplateAuctionEndingSoon EmailTemplate = "auction_ending_soon"
	TemplateAuctionWon        EmailTemplate = "auction_won"
	TemplateAuctionEnded      EmailTemplate = "auction_ended"

	// Transaction events
	TemplateTransactionCreated  EmailTemplate = "transaction_created"
	TemplateEscrowFunded        EmailTemplate = "escrow_funded"
	TemplateDeliveryConfirmed   EmailTemplate = "delivery_confirmed"
	TemplateTransactionComplete EmailTemplate = "transaction_complete"
	TemplateDisputeOpened       EmailTemplate = "dispute_opened"
)

// QueuedEmail represents an email in the queue.
type QueuedEmail struct {
	ID               uuid.UUID         `json:"id"`
	RecipientEmail   string            `json:"recipient_email"`
	RecipientAgentID uuid.UUID         `json:"recipient_agent_id"`
	Template         EmailTemplate     `json:"template"`
	Payload          map[string]any    `json:"payload"`
	Status           EmailStatus       `json:"status"`
	Attempts         int               `json:"attempts"`
	LastAttemptAt    *time.Time        `json:"last_attempt_at,omitempty"`
	SentAt           *time.Time        `json:"sent_at,omitempty"`
	Error            *string           `json:"error,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
}

// EmailStatus represents the status of a queued email.
type EmailStatus string

const (
	EmailStatusPending EmailStatus = "pending"
	EmailStatusSent    EmailStatus = "sent"
	EmailStatusFailed  EmailStatus = "failed"
)

// EmailData contains the data needed to render and send an email.
type EmailData struct {
	To          string
	ToName      string
	Subject     string
	HTMLContent string
	TextContent string
}
