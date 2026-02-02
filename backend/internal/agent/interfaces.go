package agent

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RepositoryInterface defines the contract for agent data persistence.
// This interface enables mock implementations for testing.
type RepositoryInterface interface {
	// Create inserts a new agent into the database.
	Create(ctx context.Context, agent *Agent) error

	// GetByID retrieves an agent by ID.
	GetByID(ctx context.Context, id uuid.UUID) (*Agent, error)

	// GetByAPIKeyHash retrieves an agent by API key hash.
	GetByAPIKeyHash(ctx context.Context, hash string) (*Agent, error)

	// Update updates an existing agent.
	Update(ctx context.Context, agent *Agent) error

	// UpdateLastSeen updates the agent's last seen timestamp.
	UpdateLastSeen(ctx context.Context, id uuid.UUID) error

	// Deactivate deactivates an agent (soft delete).
	Deactivate(ctx context.Context, id uuid.UUID) error

	// GetReputation retrieves the reputation details for an agent.
	GetReputation(ctx context.Context, agentID uuid.UUID) (*Reputation, error)

	// CreateOwnershipToken creates a new ownership token for an agent.
	CreateOwnershipToken(ctx context.Context, agentID uuid.UUID, tokenHash string, expiresAt time.Time) error

	// GetOwnershipTokenByHash retrieves an ownership token by its hash.
	GetOwnershipTokenByHash(ctx context.Context, tokenHash string) (*OwnershipToken, error)

	// MarkTokenUsed marks an ownership token as used by a user.
	MarkTokenUsed(ctx context.Context, tokenID, userID uuid.UUID) error

	// SetAgentOwner sets the owner of an agent.
	SetAgentOwner(ctx context.Context, agentID, userID uuid.UUID) error

	// GetAgentOwnerID retrieves the owner user ID for an agent.
	GetAgentOwnerID(ctx context.Context, agentID uuid.UUID) (*uuid.UUID, error)

	// GetAgentsByOwner retrieves all agents owned by a user.
	GetAgentsByOwner(ctx context.Context, userID uuid.UUID) ([]*Agent, error)

	// CountActiveListings counts the number of active listings for an agent.
	CountActiveListings(ctx context.Context, agentID uuid.UUID) (int, error)
}

// Verify that Repository implements RepositoryInterface
var _ RepositoryInterface = (*Repository)(nil)
