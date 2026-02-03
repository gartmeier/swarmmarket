package task

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// RepositoryInterface defines task data persistence operations.
type RepositoryInterface interface {
	// Task CRUD
	CreateTask(ctx context.Context, task *Task) error
	GetTaskByID(ctx context.Context, id uuid.UUID) (*Task, error)
	ListTasks(ctx context.Context, params ListTasksParams) (*TaskListResult, error)
	UpdateTask(ctx context.Context, task *Task) error

	// Status management
	UpdateTaskStatus(ctx context.Context, id uuid.UUID, status TaskStatus, event string, eventData json.RawMessage) error

	// History
	RecordStatusHistory(ctx context.Context, history *TaskStatusHistory) error
	GetTaskHistory(ctx context.Context, taskID uuid.UUID) ([]*TaskStatusHistory, error)

	// Transaction linking
	SetTransactionID(ctx context.Context, taskID, transactionID uuid.UUID) error
}

// CapabilityGetter retrieves capability details for validation.
type CapabilityGetter interface {
	GetCapabilityByID(ctx context.Context, id uuid.UUID) (*CapabilityInfo, error)
}

// CapabilityInfo contains the capability fields needed for task validation.
type CapabilityInfo struct {
	ID               uuid.UUID
	AgentID          uuid.UUID
	Name             string
	InputSchema      json.RawMessage
	OutputSchema     json.RawMessage
	StatusEvents     json.RawMessage
	IsActive         bool
	IsAcceptingTasks bool
	BaseFee          *float64
	PercentageFee    *float64
	Currency         string
	PricingModel     string
}

// TransactionCreator creates transactions for accepted tasks.
type TransactionCreator interface {
	CreateFromTask(ctx context.Context, requesterID, executorID uuid.UUID, taskID *uuid.UUID, amount float64, currency string) (uuid.UUID, error)
}

// SchemaValidator validates JSON data against JSON Schema.
type SchemaValidator interface {
	Validate(schema, data json.RawMessage) error
}

// EventPublisher publishes events for notifications.
type EventPublisher interface {
	Publish(ctx context.Context, eventType string, payload map[string]any) error
}

// CallbackDeliverer delivers webhooks to callback URLs.
type CallbackDeliverer interface {
	DeliverCallback(ctx context.Context, url, secret string, payload TaskCallback) error
}

// CapabilityStatsUpdater updates capability statistics after task completion.
type CapabilityStatsUpdater interface {
	RecordTaskCompletion(ctx context.Context, capabilityID uuid.UUID, success bool, rating *float64) error
}
