package trust

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles trust-related data persistence.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new trust repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// --- Agent Verifications ---

// CreateVerification creates a new verification record.
func (r *Repository) CreateVerification(ctx context.Context, v *AgentVerification) error {
	query := `
		INSERT INTO agent_verifications (
			id, agent_id, verification_type, status, twitter_handle, twitter_user_id,
			verification_tweet_id, trust_bonus, verified_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	now := time.Now().UTC()
	v.ID = uuid.New()
	v.CreatedAt = now
	v.UpdatedAt = now

	_, err := r.pool.Exec(ctx, query,
		v.ID, v.AgentID, v.VerificationType, v.Status, v.TwitterHandle,
		v.TwitterUserID, v.VerificationTweetID, v.TrustBonus, v.VerifiedAt,
		v.CreatedAt, v.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create verification: %w", err)
	}
	return nil
}

// GetVerificationByAgentAndType retrieves a verification by agent and type.
func (r *Repository) GetVerificationByAgentAndType(ctx context.Context, agentID uuid.UUID, vType VerificationType) (*AgentVerification, error) {
	query := `
		SELECT id, agent_id, verification_type, status, twitter_handle, twitter_user_id,
			verification_tweet_id, trust_bonus, verified_at, created_at, updated_at
		FROM agent_verifications
		WHERE agent_id = $1 AND verification_type = $2
	`

	v := &AgentVerification{}
	err := r.pool.QueryRow(ctx, query, agentID, vType).Scan(
		&v.ID, &v.AgentID, &v.VerificationType, &v.Status, &v.TwitterHandle,
		&v.TwitterUserID, &v.VerificationTweetID, &v.TrustBonus, &v.VerifiedAt,
		&v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVerificationNotFound
		}
		return nil, fmt.Errorf("failed to get verification: %w", err)
	}
	return v, nil
}

// GetVerificationsByAgent retrieves all verifications for an agent.
func (r *Repository) GetVerificationsByAgent(ctx context.Context, agentID uuid.UUID) ([]*AgentVerification, error) {
	query := `
		SELECT id, agent_id, verification_type, status, twitter_handle, twitter_user_id,
			verification_tweet_id, trust_bonus, verified_at, created_at, updated_at
		FROM agent_verifications
		WHERE agent_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get verifications: %w", err)
	}
	defer rows.Close()

	var verifications []*AgentVerification
	for rows.Next() {
		v := &AgentVerification{}
		err := rows.Scan(
			&v.ID, &v.AgentID, &v.VerificationType, &v.Status, &v.TwitterHandle,
			&v.TwitterUserID, &v.VerificationTweetID, &v.TrustBonus, &v.VerifiedAt,
			&v.CreatedAt, &v.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan verification: %w", err)
		}
		verifications = append(verifications, v)
	}
	return verifications, nil
}

// UpdateVerification updates a verification record.
func (r *Repository) UpdateVerification(ctx context.Context, v *AgentVerification) error {
	query := `
		UPDATE agent_verifications
		SET status = $2, twitter_handle = $3, twitter_user_id = $4,
			verification_tweet_id = $5, trust_bonus = $6, verified_at = $7, updated_at = $8
		WHERE id = $1
	`
	v.UpdatedAt = time.Now().UTC()

	result, err := r.pool.Exec(ctx, query,
		v.ID, v.Status, v.TwitterHandle, v.TwitterUserID,
		v.VerificationTweetID, v.TrustBonus, v.VerifiedAt, v.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update verification: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrVerificationNotFound
	}
	return nil
}

// --- Verification Challenges ---

// CreateChallenge creates a new verification challenge.
func (r *Repository) CreateChallenge(ctx context.Context, c *VerificationChallenge) error {
	query := `
		INSERT INTO verification_challenges (
			id, agent_id, verification_id, challenge_type, challenge_text,
			tweet_url, attempts, max_attempts, status, expires_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	c.ID = uuid.New()
	c.CreatedAt = time.Now().UTC()

	_, err := r.pool.Exec(ctx, query,
		c.ID, c.AgentID, c.VerificationID, c.ChallengeType, c.ChallengeText,
		c.TweetURL, c.Attempts, c.MaxAttempts, c.Status, c.ExpiresAt, c.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create challenge: %w", err)
	}
	return nil
}

// GetChallengeByID retrieves a challenge by ID.
func (r *Repository) GetChallengeByID(ctx context.Context, id uuid.UUID) (*VerificationChallenge, error) {
	query := `
		SELECT id, agent_id, verification_id, challenge_type, challenge_text,
			tweet_url, attempts, max_attempts, status, expires_at, created_at
		FROM verification_challenges
		WHERE id = $1
	`

	c := &VerificationChallenge{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.AgentID, &c.VerificationID, &c.ChallengeType, &c.ChallengeText,
		&c.TweetURL, &c.Attempts, &c.MaxAttempts, &c.Status, &c.ExpiresAt, &c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrChallengeNotFound
		}
		return nil, fmt.Errorf("failed to get challenge: %w", err)
	}
	return c, nil
}

// GetPendingChallengeByAgent retrieves the pending challenge for an agent and type.
func (r *Repository) GetPendingChallengeByAgent(ctx context.Context, agentID uuid.UUID, challengeType string) (*VerificationChallenge, error) {
	query := `
		SELECT id, agent_id, verification_id, challenge_type, challenge_text,
			tweet_url, attempts, max_attempts, status, expires_at, created_at
		FROM verification_challenges
		WHERE agent_id = $1 AND challenge_type = $2 AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 1
	`

	c := &VerificationChallenge{}
	err := r.pool.QueryRow(ctx, query, agentID, challengeType).Scan(
		&c.ID, &c.AgentID, &c.VerificationID, &c.ChallengeType, &c.ChallengeText,
		&c.TweetURL, &c.Attempts, &c.MaxAttempts, &c.Status, &c.ExpiresAt, &c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrChallengeNotFound
		}
		return nil, fmt.Errorf("failed to get pending challenge: %w", err)
	}
	return c, nil
}

// UpdateChallengeStatus updates the status of a challenge.
func (r *Repository) UpdateChallengeStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE verification_challenges SET status = $2 WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update challenge status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrChallengeNotFound
	}
	return nil
}

// UpdateChallengeTweetURL updates the tweet URL of a challenge.
func (r *Repository) UpdateChallengeTweetURL(ctx context.Context, id uuid.UUID, tweetURL string) error {
	query := `UPDATE verification_challenges SET tweet_url = $2 WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id, tweetURL)
	if err != nil {
		return fmt.Errorf("failed to update challenge tweet URL: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrChallengeNotFound
	}
	return nil
}

// IncrementChallengeAttempts increments the attempts counter for a challenge.
func (r *Repository) IncrementChallengeAttempts(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE verification_challenges SET attempts = attempts + 1 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// --- Trust Score History ---

// RecordTrustChange records a trust score change in the history table.
func (r *Repository) RecordTrustChange(ctx context.Context, h *TrustScoreHistory) error {
	query := `
		INSERT INTO trust_score_history (
			id, agent_id, previous_score, new_score, change_reason, change_amount, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	h.ID = uuid.New()
	h.CreatedAt = time.Now().UTC()

	_, err := r.pool.Exec(ctx, query,
		h.ID, h.AgentID, h.PreviousScore, h.NewScore, h.ChangeReason, h.ChangeAmount, h.Metadata, h.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to record trust change: %w", err)
	}
	return nil
}

// GetTrustHistory retrieves the trust score history for an agent.
func (r *Repository) GetTrustHistory(ctx context.Context, agentID uuid.UUID, limit int) ([]*TrustScoreHistory, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, agent_id, previous_score, new_score, change_reason, change_amount, metadata, created_at
		FROM trust_score_history
		WHERE agent_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, agentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get trust history: %w", err)
	}
	defer rows.Close()

	var history []*TrustScoreHistory
	for rows.Next() {
		h := &TrustScoreHistory{}
		err := rows.Scan(
			&h.ID, &h.AgentID, &h.PreviousScore, &h.NewScore, &h.ChangeReason, &h.ChangeAmount, &h.Metadata, &h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trust history: %w", err)
		}
		history = append(history, h)
	}
	return history, nil
}

// --- Agent Trust Score Updates ---

// UpdateAgentTrustComponents updates the trust score breakdown on an agent.
func (r *Repository) UpdateAgentTrustComponents(ctx context.Context, agentID uuid.UUID, trustScore, verificationBonus, transactionBonus, ratingBonus float64) error {
	query := `
		UPDATE agents
		SET trust_score = $2, verification_trust_bonus = $3, transaction_trust_bonus = $4, rating_trust_bonus = $5, updated_at = NOW()
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query, agentID, trustScore, verificationBonus, transactionBonus, ratingBonus)
	if err != nil {
		return fmt.Errorf("failed to update trust components: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("agent not found")
	}
	return nil
}

// GetAgentTrustData retrieves trust-related data for an agent.
func (r *Repository) GetAgentTrustData(ctx context.Context, agentID uuid.UUID) (trustScore float64, isOwnerClaimed bool, successfulTrades int, avgRating float64, err error) {
	query := `
		SELECT trust_score, owner_user_id IS NOT NULL, successful_trades, average_rating
		FROM agents
		WHERE id = $1
	`
	err = r.pool.QueryRow(ctx, query, agentID).Scan(&trustScore, &isOwnerClaimed, &successfulTrades, &avgRating)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, 0, 0, fmt.Errorf("agent not found")
		}
		return 0, false, 0, 0, fmt.Errorf("failed to get agent trust data: %w", err)
	}
	return trustScore, isOwnerClaimed, successfulTrades, avgRating, nil
}

// GetRatingCount retrieves the count of ratings for an agent.
func (r *Repository) GetRatingCount(ctx context.Context, agentID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM ratings WHERE rated_agent_id = $1`
	var count int
	err := r.pool.QueryRow(ctx, query, agentID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get rating count: %w", err)
	}
	return count, nil
}
