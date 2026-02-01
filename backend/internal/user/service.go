package user

import (
	"context"

	"github.com/google/uuid"
)

// ClerkUser represents user data from Clerk.
type ClerkUser struct {
	ID            string
	EmailAddress  string
	FirstName     string
	LastName      string
	ImageURL      string
}

// Service handles user business logic.
type Service struct {
	repo *Repository
}

// NewService creates a new user service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetOrCreateUser upserts a user from Clerk data.
func (s *Service) GetOrCreateUser(ctx context.Context, clerkUser *ClerkUser) (*User, error) {
	name := clerkUser.FirstName
	if clerkUser.LastName != "" {
		if name != "" {
			name += " "
		}
		name += clerkUser.LastName
	}

	user := &User{
		ClerkUserID: clerkUser.ID,
		Email:       clerkUser.EmailAddress,
		Name:        name,
		AvatarURL:   clerkUser.ImageURL,
	}

	return s.repo.UpsertUser(ctx, user)
}

// GetUserByID retrieves a user by internal ID.
func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.GetUserByID(ctx, id)
}

// GetUserByClerkID retrieves a user by Clerk ID.
func (s *Service) GetUserByClerkID(ctx context.Context, clerkUserID string) (*User, error) {
	return s.repo.GetUserByClerkID(ctx, clerkUserID)
}

// GetOwnedAgents retrieves all agents owned by a user.
func (s *Service) GetOwnedAgents(ctx context.Context, userID uuid.UUID) ([]*OwnedAgentSummary, error) {
	return s.repo.GetOwnedAgents(ctx, userID)
}

// GetAgentMetrics retrieves detailed metrics for an owned agent.
func (s *Service) GetAgentMetrics(ctx context.Context, userID, agentID uuid.UUID) (*AgentMetrics, error) {
	return s.repo.GetAgentMetrics(ctx, userID, agentID)
}
