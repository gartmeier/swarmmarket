package task

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the core status of a task.
type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"     // Created, waiting for executor acceptance
	StatusAccepted   TaskStatus = "accepted"    // Executor accepted, awaiting payment/start
	StatusInProgress TaskStatus = "in_progress" // Work has begun
	StatusDelivered  TaskStatus = "delivered"   // Executor claims completion with output
	StatusCompleted  TaskStatus = "completed"   // Requester confirmed, task done
	StatusCancelled  TaskStatus = "cancelled"   // Cancelled by requester
	StatusFailed     TaskStatus = "failed"      // Failed permanently
)

// IsTerminal returns true if the status is a terminal state.
func (s TaskStatus) IsTerminal() bool {
	return s == StatusCompleted || s == StatusCancelled || s == StatusFailed
}

// Task represents a capability-linked unit of work.
type Task struct {
	ID           uuid.UUID `json:"id" db:"id"`
	RequesterID  uuid.UUID `json:"requester_id" db:"requester_id"`
	ExecutorID   uuid.UUID `json:"executor_id" db:"executor_id"`
	CapabilityID uuid.UUID `json:"capability_id" db:"capability_id"`

	// Validated input/output
	Input  json.RawMessage `json:"input" db:"input"`
	Output json.RawMessage `json:"output,omitempty" db:"output"`

	// Status
	Status           TaskStatus      `json:"status" db:"status"`
	CurrentEvent     string          `json:"current_event,omitempty" db:"current_event"`
	CurrentEventData json.RawMessage `json:"current_event_data,omitempty" db:"current_event_data"`

	// Callback
	CallbackURL    string `json:"callback_url,omitempty" db:"callback_url"`
	CallbackSecret string `json:"-" db:"callback_secret"` // Never expose in JSON

	// Pricing
	PriceAmount   float64 `json:"price_amount" db:"price_amount"`
	PriceCurrency string  `json:"price_currency" db:"price_currency"`

	// Linked transaction
	TransactionID *uuid.UUID `json:"transaction_id,omitempty" db:"transaction_id"`

	// Error handling
	ErrorMessage string `json:"error_message,omitempty" db:"error_message"`
	RetryCount   int    `json:"retry_count" db:"retry_count"`
	MaxRetries   int    `json:"max_retries" db:"max_retries"`

	// Timestamps
	DeadlineAt  *time.Time `json:"deadline_at,omitempty" db:"deadline_at"`
	StartedAt   *time.Time `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`

	// Metadata
	Metadata  map[string]any `json:"metadata,omitempty" db:"metadata"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`

	// Joined data (not stored in tasks table)
	RequesterName  string `json:"requester_name,omitempty" db:"requester_name"`
	ExecutorName   string `json:"executor_name,omitempty" db:"executor_name"`
	CapabilityName string `json:"capability_name,omitempty" db:"capability_name"`
}

// TaskStatusHistory records state transitions for audit.
type TaskStatusHistory struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	TaskID     uuid.UUID       `json:"task_id" db:"task_id"`
	FromStatus *TaskStatus     `json:"from_status,omitempty" db:"from_status"`
	ToStatus   TaskStatus      `json:"to_status" db:"to_status"`
	Event      string          `json:"event,omitempty" db:"event"`
	EventData  json.RawMessage `json:"event_data,omitempty" db:"event_data"`
	ChangedBy  *uuid.UUID      `json:"changed_by,omitempty" db:"changed_by"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

// --- Request/Response DTOs ---

// CreateTaskRequest is the request to create a new task.
type CreateTaskRequest struct {
	CapabilityID   uuid.UUID       `json:"capability_id"`
	Input          json.RawMessage `json:"input"`
	CallbackURL    string          `json:"callback_url,omitempty"`
	CallbackSecret string          `json:"callback_secret,omitempty"`
	DeadlineAt     *time.Time      `json:"deadline_at,omitempty"`
	Metadata       map[string]any  `json:"metadata,omitempty"`
}

// UpdateTaskProgressRequest is the request to update task progress with custom events.
type UpdateTaskProgressRequest struct {
	Event     string          `json:"event,omitempty"`
	EventData json.RawMessage `json:"event_data,omitempty"`
	Message   string          `json:"message,omitempty"`
}

// DeliverTaskRequest is the request to deliver task output.
type DeliverTaskRequest struct {
	Output json.RawMessage `json:"output"`
}

// FailTaskRequest is the request to mark a task as failed.
type FailTaskRequest struct {
	ErrorMessage string `json:"error_message"`
	Retry        bool   `json:"retry,omitempty"`
}

// ListTasksParams contains filter parameters for listing tasks.
type ListTasksParams struct {
	RequesterID  *uuid.UUID
	ExecutorID   *uuid.UUID
	CapabilityID *uuid.UUID
	Status       *TaskStatus
	AgentID      *uuid.UUID // Filter where agent is requester OR executor
	Limit        int
	Offset       int
}

// TaskListResult is a paginated list of tasks.
type TaskListResult struct {
	Items  []*Task `json:"items"`
	Total  int     `json:"total"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
}

// TaskCallback is the webhook payload sent to callback_url.
type TaskCallback struct {
	TaskID        uuid.UUID       `json:"task_id"`
	CapabilityID  uuid.UUID       `json:"capability_id"`
	Status        TaskStatus      `json:"status"`
	Event         string          `json:"event,omitempty"`
	EventData     json.RawMessage `json:"event_data,omitempty"`
	Output        json.RawMessage `json:"output,omitempty"`
	Error         string          `json:"error,omitempty"`
	TransactionID *uuid.UUID      `json:"transaction_id,omitempty"`
	Timestamp     time.Time       `json:"timestamp"`
}
