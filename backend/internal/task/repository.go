package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository implements RepositoryInterface using PostgreSQL.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new task repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// CreateTask inserts a new task into the database.
func (r *Repository) CreateTask(ctx context.Context, task *Task) error {
	query := `
		INSERT INTO tasks (
			id, requester_id, executor_id, capability_id,
			input, status, callback_url, callback_secret,
			price_amount, price_currency, deadline_at, metadata,
			max_retries, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	metadataJSON, _ := json.Marshal(task.Metadata)
	if task.Metadata == nil {
		metadataJSON = []byte("{}")
	}

	_, err := r.pool.Exec(ctx, query,
		task.ID,
		task.RequesterID,
		task.ExecutorID,
		task.CapabilityID,
		task.Input,
		task.Status,
		task.CallbackURL,
		task.CallbackSecret,
		task.PriceAmount,
		task.PriceCurrency,
		task.DeadlineAt,
		metadataJSON,
		task.MaxRetries,
		task.CreatedAt,
		task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

// GetTaskByID retrieves a task by its ID with joined agent and capability names.
func (r *Repository) GetTaskByID(ctx context.Context, id uuid.UUID) (*Task, error) {
	query := `
		SELECT
			t.id, t.requester_id, t.executor_id, t.capability_id,
			t.input, t.output, t.status, t.current_event, t.current_event_data,
			t.callback_url, t.callback_secret,
			t.price_amount, t.price_currency, t.transaction_id,
			t.error_message, t.retry_count, t.max_retries,
			t.deadline_at, t.started_at, t.completed_at,
			t.metadata, t.created_at, t.updated_at,
			req.name as requester_name,
			exec.name as executor_name,
			c.name as capability_name
		FROM tasks t
		LEFT JOIN agents req ON t.requester_id = req.id
		LEFT JOIN agents exec ON t.executor_id = exec.id
		LEFT JOIN capabilities c ON t.capability_id = c.id
		WHERE t.id = $1
	`

	var task Task
	var metadataJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&task.ID, &task.RequesterID, &task.ExecutorID, &task.CapabilityID,
		&task.Input, &task.Output, &task.Status, &task.CurrentEvent, &task.CurrentEventData,
		&task.CallbackURL, &task.CallbackSecret,
		&task.PriceAmount, &task.PriceCurrency, &task.TransactionID,
		&task.ErrorMessage, &task.RetryCount, &task.MaxRetries,
		&task.DeadlineAt, &task.StartedAt, &task.CompletedAt,
		&metadataJSON, &task.CreatedAt, &task.UpdatedAt,
		&task.RequesterName, &task.ExecutorName, &task.CapabilityName,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &task.Metadata)
	}

	return &task, nil
}

// ListTasks retrieves tasks with filters and pagination.
func (r *Repository) ListTasks(ctx context.Context, params ListTasksParams) (*TaskListResult, error) {
	var conditions []string
	var args []any
	argNum := 1

	// Build WHERE conditions
	if params.RequesterID != nil && params.ExecutorID != nil {
		// OR condition: agent is either requester or executor
		conditions = append(conditions, fmt.Sprintf("(t.requester_id = $%d OR t.executor_id = $%d)", argNum, argNum+1))
		args = append(args, *params.RequesterID, *params.ExecutorID)
		argNum += 2
	} else {
		if params.RequesterID != nil {
			conditions = append(conditions, fmt.Sprintf("t.requester_id = $%d", argNum))
			args = append(args, *params.RequesterID)
			argNum++
		}
		if params.ExecutorID != nil {
			conditions = append(conditions, fmt.Sprintf("t.executor_id = $%d", argNum))
			args = append(args, *params.ExecutorID)
			argNum++
		}
	}

	if params.AgentID != nil {
		conditions = append(conditions, fmt.Sprintf("(t.requester_id = $%d OR t.executor_id = $%d)", argNum, argNum+1))
		args = append(args, *params.AgentID, *params.AgentID)
		argNum += 2
	}

	if params.CapabilityID != nil {
		conditions = append(conditions, fmt.Sprintf("t.capability_id = $%d", argNum))
		args = append(args, *params.CapabilityID)
		argNum++
	}

	if params.Status != nil {
		conditions = append(conditions, fmt.Sprintf("t.status = $%d", argNum))
		args = append(args, *params.Status)
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks t %s", whereClause)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count tasks: %w", err)
	}

	// Fetch items
	selectQuery := fmt.Sprintf(`
		SELECT
			t.id, t.requester_id, t.executor_id, t.capability_id,
			t.input, t.output, t.status, t.current_event, t.current_event_data,
			t.callback_url,
			t.price_amount, t.price_currency, t.transaction_id,
			t.error_message, t.retry_count, t.max_retries,
			t.deadline_at, t.started_at, t.completed_at,
			t.metadata, t.created_at, t.updated_at,
			req.name as requester_name,
			exec.name as executor_name,
			c.name as capability_name
		FROM tasks t
		LEFT JOIN agents req ON t.requester_id = req.id
		LEFT JOIN agents exec ON t.executor_id = exec.id
		LEFT JOIN capabilities c ON t.capability_id = c.id
		%s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)

	args = append(args, params.Limit, params.Offset)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var task Task
		var metadataJSON []byte

		err := rows.Scan(
			&task.ID, &task.RequesterID, &task.ExecutorID, &task.CapabilityID,
			&task.Input, &task.Output, &task.Status, &task.CurrentEvent, &task.CurrentEventData,
			&task.CallbackURL,
			&task.PriceAmount, &task.PriceCurrency, &task.TransactionID,
			&task.ErrorMessage, &task.RetryCount, &task.MaxRetries,
			&task.DeadlineAt, &task.StartedAt, &task.CompletedAt,
			&metadataJSON, &task.CreatedAt, &task.UpdatedAt,
			&task.RequesterName, &task.ExecutorName, &task.CapabilityName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &task.Metadata)
		}

		tasks = append(tasks, &task)
	}

	return &TaskListResult{
		Items:  tasks,
		Total:  total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

// UpdateTask updates a task's mutable fields.
func (r *Repository) UpdateTask(ctx context.Context, task *Task) error {
	query := `
		UPDATE tasks SET
			output = $2,
			status = $3,
			current_event = $4,
			current_event_data = $5,
			error_message = $6,
			retry_count = $7,
			started_at = $8,
			completed_at = $9,
			metadata = $10,
			updated_at = NOW()
		WHERE id = $1
	`

	metadataJSON, _ := json.Marshal(task.Metadata)
	if task.Metadata == nil {
		metadataJSON = []byte("{}")
	}

	result, err := r.pool.Exec(ctx, query,
		task.ID,
		task.Output,
		task.Status,
		task.CurrentEvent,
		task.CurrentEventData,
		task.ErrorMessage,
		task.RetryCount,
		task.StartedAt,
		task.CompletedAt,
		metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// UpdateTaskStatus updates just the status and event fields.
func (r *Repository) UpdateTaskStatus(ctx context.Context, id uuid.UUID, status TaskStatus, event string, eventData json.RawMessage) error {
	query := `
		UPDATE tasks SET
			status = $2,
			current_event = $3,
			current_event_data = $4,
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id, status, event, eventData)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// RecordStatusHistory inserts a status change record.
func (r *Repository) RecordStatusHistory(ctx context.Context, history *TaskStatusHistory) error {
	if history.ID == uuid.Nil {
		history.ID = uuid.New()
	}
	if history.CreatedAt.IsZero() {
		history.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO task_status_history (
			id, task_id, from_status, to_status, event, event_data, changed_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		history.ID,
		history.TaskID,
		history.FromStatus,
		history.ToStatus,
		history.Event,
		history.EventData,
		history.ChangedBy,
		history.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to record status history: %w", err)
	}

	return nil
}

// GetTaskHistory retrieves the status history for a task.
func (r *Repository) GetTaskHistory(ctx context.Context, taskID uuid.UUID) ([]*TaskStatusHistory, error) {
	query := `
		SELECT id, task_id, from_status, to_status, event, event_data, changed_by, created_at
		FROM task_status_history
		WHERE task_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task history: %w", err)
	}
	defer rows.Close()

	var history []*TaskStatusHistory
	for rows.Next() {
		var h TaskStatusHistory
		err := rows.Scan(
			&h.ID, &h.TaskID, &h.FromStatus, &h.ToStatus,
			&h.Event, &h.EventData, &h.ChangedBy, &h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history: %w", err)
		}
		history = append(history, &h)
	}

	return history, nil
}

// SetTransactionID links a task to a transaction.
func (r *Repository) SetTransactionID(ctx context.Context, taskID, transactionID uuid.UUID) error {
	query := `UPDATE tasks SET transaction_id = $2, updated_at = NOW() WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, taskID, transactionID)
	if err != nil {
		return fmt.Errorf("failed to set transaction id: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}
