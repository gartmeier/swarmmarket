package agent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAgentNotFound       = errors.New("agent not found")
	ErrDuplicateKey        = errors.New("duplicate api key")
	ErrTokenNotFound       = errors.New("ownership token not found")
	ErrTokenExpired        = errors.New("ownership token expired")
	ErrTokenAlreadyUsed    = errors.New("ownership token already used")
	ErrAgentAlreadyOwned   = errors.New("agent already has an owner")
)

// OwnershipToken represents a token for claiming agent ownership.
type OwnershipToken struct {
	ID           uuid.UUID
	AgentID      uuid.UUID
	TokenHash    string
	ExpiresAt    time.Time
	UsedAt       *time.Time
	UsedByUserID *uuid.UUID
	CreatedAt    time.Time
}

// Repository handles agent data persistence.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new agent repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new agent into the database.
func (r *Repository) Create(ctx context.Context, agent *Agent) error {
	query := `
		INSERT INTO agents (
			id, name, description, avatar_url, owner_email, api_key_hash, api_key_prefix,
			verification_level, trust_score, total_transactions, successful_trades,
			average_rating, is_active, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`

	_, err := r.pool.Exec(ctx, query,
		agent.ID,
		agent.Name,
		agent.Description,
		agent.AvatarURL,
		agent.OwnerEmail,
		agent.APIKeyHash,
		agent.APIKeyPrefix,
		agent.VerificationLevel,
		agent.TrustScore,
		agent.TotalTransactions,
		agent.SuccessfulTrades,
		agent.AverageRating,
		agent.IsActive,
		agent.Metadata,
		agent.CreatedAt,
		agent.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	return nil
}

// GetByID retrieves an agent by ID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Agent, error) {
	query := `
		SELECT id, name, description, avatar_url, owner_email, owner_user_id, api_key_hash, api_key_prefix,
			verification_level, trust_score, total_transactions, successful_trades,
			average_rating, is_active, metadata, created_at, updated_at, last_seen_at
		FROM agents
		WHERE id = $1
	`

	agent := &Agent{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&agent.ID,
		&agent.Name,
		&agent.Description,
		&agent.AvatarURL,
		&agent.OwnerEmail,
		&agent.OwnerUserID,
		&agent.APIKeyHash,
		&agent.APIKeyPrefix,
		&agent.VerificationLevel,
		&agent.TrustScore,
		&agent.TotalTransactions,
		&agent.SuccessfulTrades,
		&agent.AverageRating,
		&agent.IsActive,
		&agent.Metadata,
		&agent.CreatedAt,
		&agent.UpdatedAt,
		&agent.LastSeenAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAgentNotFound
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	return agent, nil
}

// GetByAPIKeyHash retrieves an agent by API key hash.
func (r *Repository) GetByAPIKeyHash(ctx context.Context, hash string) (*Agent, error) {
	query := `
		SELECT id, name, description, avatar_url, owner_email, owner_user_id, api_key_hash, api_key_prefix,
			verification_level, trust_score, total_transactions, successful_trades,
			average_rating, is_active, metadata, created_at, updated_at, last_seen_at
		FROM agents
		WHERE api_key_hash = $1 AND is_active = true
	`

	agent := &Agent{}
	err := r.pool.QueryRow(ctx, query, hash).Scan(
		&agent.ID,
		&agent.Name,
		&agent.Description,
		&agent.AvatarURL,
		&agent.OwnerEmail,
		&agent.OwnerUserID,
		&agent.APIKeyHash,
		&agent.APIKeyPrefix,
		&agent.VerificationLevel,
		&agent.TrustScore,
		&agent.TotalTransactions,
		&agent.SuccessfulTrades,
		&agent.AverageRating,
		&agent.IsActive,
		&agent.Metadata,
		&agent.CreatedAt,
		&agent.UpdatedAt,
		&agent.LastSeenAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAgentNotFound
		}
		return nil, fmt.Errorf("failed to get agent by api key: %w", err)
	}

	return agent, nil
}

// Update updates an existing agent.
func (r *Repository) Update(ctx context.Context, agent *Agent) error {
	query := `
		UPDATE agents
		SET name = $2, description = $3, avatar_url = $4, metadata = $5, updated_at = $6
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		agent.ID,
		agent.Name,
		agent.Description,
		agent.AvatarURL,
		agent.Metadata,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("failed to update agent: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrAgentNotFound
	}

	return nil
}

// UpdateLastSeen updates the agent's last seen timestamp.
func (r *Repository) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE agents SET last_seen_at = $2 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, time.Now().UTC())
	return err
}

// Deactivate deactivates an agent (soft delete).
func (r *Repository) Deactivate(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE agents SET is_active = false, updated_at = $2 WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to deactivate agent: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrAgentNotFound
	}

	return nil
}

// GetReputation retrieves the reputation details for an agent.
func (r *Repository) GetReputation(ctx context.Context, agentID uuid.UUID) (*Reputation, error) {
	// First verify agent exists
	agent, err := r.GetByID(ctx, agentID)
	if err != nil {
		return nil, err
	}

	rep := &Reputation{
		AgentID:           agent.ID,
		TrustScore:        agent.TrustScore,
		TotalTransactions: agent.TotalTransactions,
		SuccessfulTrades:  agent.SuccessfulTrades,
		AverageRating:     agent.AverageRating,
	}

	// Get recent ratings
	ratingsQuery := `
		SELECT transaction_id, rater_id, score, comment, created_at
		FROM ratings
		WHERE rated_agent_id = $1
		ORDER BY created_at DESC
		LIMIT 10
	`

	rows, err := r.pool.Query(ctx, ratingsQuery, agentID)
	if err != nil {
		// Ratings table might not exist yet, continue without ratings
		return rep, nil
	}
	defer rows.Close()

	for rows.Next() {
		var rating Rating
		if err := rows.Scan(
			&rating.TransactionID,
			&rating.RaterID,
			&rating.Score,
			&rating.Comment,
			&rating.CreatedAt,
		); err != nil {
			continue
		}
		rep.RecentRatings = append(rep.RecentRatings, rating)
	}

	rep.RatingCount = len(rep.RecentRatings)

	return rep, nil
}

// CreateOwnershipToken creates a new ownership token for an agent.
func (r *Repository) CreateOwnershipToken(ctx context.Context, agentID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO agent_ownership_tokens (agent_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.pool.Exec(ctx, query, agentID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create ownership token: %w", err)
	}
	return nil
}

// GetOwnershipTokenByHash retrieves an ownership token by its hash.
func (r *Repository) GetOwnershipTokenByHash(ctx context.Context, tokenHash string) (*OwnershipToken, error) {
	query := `
		SELECT id, agent_id, token_hash, expires_at, used_at, used_by_user_id, created_at
		FROM agent_ownership_tokens
		WHERE token_hash = $1
	`

	var token OwnershipToken
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.AgentID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.UsedAt,
		&token.UsedByUserID,
		&token.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get ownership token: %w", err)
	}

	return &token, nil
}

// MarkTokenUsed marks an ownership token as used by a user.
func (r *Repository) MarkTokenUsed(ctx context.Context, tokenID, userID uuid.UUID) error {
	query := `
		UPDATE agent_ownership_tokens
		SET used_at = NOW(), used_by_user_id = $2
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query, tokenID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrTokenNotFound
	}
	return nil
}

// SetAgentOwner sets the owner of an agent and boosts trust score to 1.0.
func (r *Repository) SetAgentOwner(ctx context.Context, agentID, userID uuid.UUID) error {
	query := `
		UPDATE agents
		SET owner_user_id = $2, trust_score = 1.0, updated_at = NOW()
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query, agentID, userID)
	if err != nil {
		return fmt.Errorf("failed to set agent owner: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrAgentNotFound
	}
	return nil
}

// GetAgentOwnerID retrieves the owner user ID for an agent.
func (r *Repository) GetAgentOwnerID(ctx context.Context, agentID uuid.UUID) (*uuid.UUID, error) {
	query := `SELECT owner_user_id FROM agents WHERE id = $1`
	var ownerID *uuid.UUID
	err := r.pool.QueryRow(ctx, query, agentID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAgentNotFound
		}
		return nil, fmt.Errorf("failed to get agent owner: %w", err)
	}
	return ownerID, nil
}

// GetAgentsByOwner retrieves all agents owned by a user.
func (r *Repository) GetAgentsByOwner(ctx context.Context, userID uuid.UUID) ([]*Agent, error) {
	query := `
		SELECT id, name, description, avatar_url, owner_email, owner_user_id, api_key_hash, api_key_prefix,
			verification_level, trust_score, total_transactions, successful_trades,
			average_rating, is_active, metadata, created_at, updated_at, last_seen_at
		FROM agents
		WHERE owner_user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agents by owner: %w", err)
	}
	defer rows.Close()

	var agents []*Agent
	for rows.Next() {
		agent := &Agent{}
		err := rows.Scan(
			&agent.ID,
			&agent.Name,
			&agent.Description,
			&agent.AvatarURL,
			&agent.OwnerEmail,
			&agent.OwnerUserID,
			&agent.APIKeyHash,
			&agent.APIKeyPrefix,
			&agent.VerificationLevel,
			&agent.TrustScore,
			&agent.TotalTransactions,
			&agent.SuccessfulTrades,
			&agent.AverageRating,
			&agent.IsActive,
			&agent.Metadata,
			&agent.CreatedAt,
			&agent.UpdatedAt,
			&agent.LastSeenAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

// CountActiveListings counts the number of active listings for an agent.
func (r *Repository) CountActiveListings(ctx context.Context, agentID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM listings WHERE seller_id = $1 AND status = 'active'`
	var count int
	err := r.pool.QueryRow(ctx, query, agentID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active listings: %w", err)
	}
	return count, nil
}
