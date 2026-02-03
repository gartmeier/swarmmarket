package trust

import (
	"time"

	"github.com/google/uuid"
)

// VerificationType represents the type of verification
type VerificationType string

const (
	VerificationTwitter   VerificationType = "twitter"
	VerificationOwnership VerificationType = "ownership"
)

// VerificationStatus represents the status of a verification
type VerificationStatus string

const (
	StatusPending  VerificationStatus = "pending"
	StatusVerified VerificationStatus = "verified"
	StatusFailed   VerificationStatus = "failed"
	StatusExpired  VerificationStatus = "expired"
)

// ChangeReason represents the reason for a trust score change
type ChangeReason string

const (
	ReasonTwitterVerified       ChangeReason = "twitter_verified"
	ReasonTransactionCompleted  ChangeReason = "transaction_completed"
	ReasonRatingReceived        ChangeReason = "rating_received"
	ReasonOwnershipClaimed      ChangeReason = "ownership_claimed"
)

// AgentVerification represents a completed or pending verification
type AgentVerification struct {
	ID                  uuid.UUID          `json:"id"`
	AgentID             uuid.UUID          `json:"agent_id"`
	VerificationType    VerificationType   `json:"verification_type"`
	Status              VerificationStatus `json:"status"`
	TwitterHandle       string             `json:"twitter_handle,omitempty"`
	TwitterUserID       string             `json:"-"`
	VerificationTweetID string             `json:"-"`
	TrustBonus          float64            `json:"trust_bonus"`
	VerifiedAt          *time.Time         `json:"verified_at,omitempty"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
}

// VerificationChallenge represents a pending verification challenge
type VerificationChallenge struct {
	ID             uuid.UUID  `json:"id"`
	AgentID        uuid.UUID  `json:"agent_id"`
	VerificationID *uuid.UUID `json:"verification_id,omitempty"`
	ChallengeType  string     `json:"challenge_type"`
	ChallengeText  string     `json:"challenge_text,omitempty"`
	TweetURL       string     `json:"tweet_url,omitempty"`
	Attempts       int        `json:"attempts"`
	MaxAttempts    int        `json:"max_attempts"`
	Status         string     `json:"status"`
	ExpiresAt      time.Time  `json:"expires_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

// TrustScoreHistory represents a change in trust score
type TrustScoreHistory struct {
	ID            uuid.UUID      `json:"id"`
	AgentID       uuid.UUID      `json:"agent_id"`
	PreviousScore float64        `json:"previous_score"`
	NewScore      float64        `json:"new_score"`
	ChangeReason  ChangeReason   `json:"change_reason"`
	ChangeAmount  float64        `json:"change_amount"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

// TrustBreakdown shows the components of an agent's trust score (0-100%)
type TrustBreakdown struct {
	AgentID           uuid.UUID             `json:"agent_id"`
	TotalScore        float64               `json:"total_score"`         // 0.0-1.0 (display as 0-100%)
	BaseScore         float64               `json:"base_score"`          // 0% for new agents
	VerificationBonus float64               `json:"verification_bonus"`  // e.g., +15% for Twitter
	TransactionBonus  float64               `json:"transaction_bonus"`   // up to +75% from trades
	HumanLinkBonus    float64               `json:"human_link_bonus"`    // +10% if linked to human
	IsOwnerClaimed    bool                  `json:"is_owner_claimed"`
	Verifications     []VerificationSummary `json:"verifications"`
	TransactionCount  int                   `json:"transaction_count"`
	SuccessfulTrades  int                   `json:"successful_trades"`
}

// VerificationSummary is a short summary of a verification
type VerificationSummary struct {
	Type       VerificationType   `json:"type"`
	Status     VerificationStatus `json:"status"`
	TrustBonus float64            `json:"trust_bonus"`
	VerifiedAt *time.Time         `json:"verified_at,omitempty"`
	Handle     string             `json:"handle,omitempty"` // Twitter handle if applicable
}

// Request/Response DTOs

// InitiateTwitterVerificationResponse is returned when starting Twitter verification
type InitiateTwitterVerificationResponse struct {
	ChallengeID   string    `json:"challenge_id"`
	ChallengeText string    `json:"challenge_text"`
	Instructions  string    `json:"instructions"`
	ExpiresAt     time.Time `json:"expires_at"`
}

// ConfirmTwitterVerificationRequest is the request to confirm Twitter verification
type ConfirmTwitterVerificationRequest struct {
	ChallengeID string `json:"challenge_id"`
	TweetURL    string `json:"tweet_url"`
}

// ConfirmVerificationResponse is returned after confirming any verification
type ConfirmVerificationResponse struct {
	Verified      bool    `json:"verified"`
	TrustBonus    float64 `json:"trust_bonus,omitempty"`
	NewTrustScore float64 `json:"new_trust_score,omitempty"`
	Message       string  `json:"message"`
}
