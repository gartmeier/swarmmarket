package messaging

import (
	"context"

	"github.com/digi604/swarmmarket/backend/internal/email"
	"github.com/google/uuid"
)

// EmailAdapter adapts the email service to the EmailQueuer interface.
type EmailAdapter struct {
	svc *email.Service
}

// NewEmailAdapter creates a new email adapter.
func NewEmailAdapter(svc *email.Service) *EmailAdapter {
	if svc == nil {
		return nil
	}
	return &EmailAdapter{svc: svc}
}

// QueueEmail queues an email for async delivery.
func (a *EmailAdapter) QueueEmail(ctx context.Context, recipientEmail string, recipientAgentID uuid.UUID, template string, payload map[string]any) error {
	return a.svc.QueueEmail(ctx, recipientEmail, recipientAgentID, email.EmailTemplate(template), payload)
}

// AgentServiceInterface defines the methods needed from agent service.
type AgentServiceInterface interface {
	GetAgentByID(ctx context.Context, id uuid.UUID) (AgentInfo, error)
}

// AgentAdapter wraps an agent service to implement AgentGetter.
type AgentAdapter struct {
	getter AgentGetterFunc
}

// AgentGetterFunc is a function type for getting agents.
type AgentGetterFunc func(ctx context.Context, id uuid.UUID) (AgentInfo, error)

// NewAgentAdapter creates a new agent adapter.
func NewAgentAdapter(getter AgentGetterFunc) *AgentAdapter {
	if getter == nil {
		return nil
	}
	return &AgentAdapter{getter: getter}
}

// GetAgentByID gets an agent by ID.
func (a *AgentAdapter) GetAgentByID(ctx context.Context, id uuid.UUID) (AgentInfo, error) {
	return a.getter(ctx, id)
}
