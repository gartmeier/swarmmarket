package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Errors
var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrNotAuthorized      = errors.New("not authorized to perform this action")
	ErrInvalidStatus      = errors.New("invalid status transition")
	ErrCapabilityNotFound = errors.New("capability not found")
	ErrCapabilityInactive = errors.New("capability is not accepting tasks")
	ErrInputValidation    = errors.New("input validation failed")
	ErrOutputValidation   = errors.New("output validation failed")
	ErrInvalidEvent       = errors.New("invalid status event for this capability")
	ErrSelfAssignment     = errors.New("cannot create task for your own capability")
)

// Service handles task business logic.
type Service struct {
	repo       RepositoryInterface
	capability CapabilityGetter
	validator  SchemaValidator
	publisher  EventPublisher
	callback   CallbackDeliverer
	txCreator  TransactionCreator
	capStats   CapabilityStatsUpdater
}

// NewService creates a new task service.
func NewService(repo RepositoryInterface, capability CapabilityGetter, publisher EventPublisher) *Service {
	return &Service{
		repo:       repo,
		capability: capability,
		publisher:  publisher,
	}
}

// SetSchemaValidator sets the JSON Schema validator.
func (s *Service) SetSchemaValidator(v SchemaValidator) {
	s.validator = v
}

// SetCallbackDeliverer sets the callback deliverer.
func (s *Service) SetCallbackDeliverer(c CallbackDeliverer) {
	s.callback = c
}

// SetTransactionCreator sets the transaction creator.
func (s *Service) SetTransactionCreator(tc TransactionCreator) {
	s.txCreator = tc
}

// SetCapabilityStatsUpdater sets the capability stats updater.
func (s *Service) SetCapabilityStatsUpdater(csu CapabilityStatsUpdater) {
	s.capStats = csu
}

// CreateTask creates a new task for a capability.
func (s *Service) CreateTask(ctx context.Context, requesterID uuid.UUID, req *CreateTaskRequest) (*Task, error) {
	// 1. Validate request
	if req.CapabilityID == uuid.Nil {
		return nil, fmt.Errorf("capability_id is required")
	}
	if len(req.Input) == 0 {
		return nil, fmt.Errorf("input is required")
	}

	// 2. Get capability
	cap, err := s.capability.GetCapabilityByID(ctx, req.CapabilityID)
	if err != nil {
		return nil, ErrCapabilityNotFound
	}

	// 3. Verify capability is active and accepting tasks
	if !cap.IsActive || !cap.IsAcceptingTasks {
		return nil, ErrCapabilityInactive
	}

	// 4. Prevent self-assignment
	if cap.AgentID == requesterID {
		return nil, ErrSelfAssignment
	}

	// 5. Validate input against capability's input_schema
	if s.validator != nil && len(cap.InputSchema) > 0 {
		if err := s.validator.Validate(cap.InputSchema, req.Input); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInputValidation, err)
		}
	}

	// 6. Calculate price
	price := s.calculatePrice(cap, req.Input)
	currency := cap.Currency
	if currency == "" {
		currency = "USD"
	}

	// 7. Create task
	now := time.Now().UTC()
	task := &Task{
		ID:             uuid.New(),
		RequesterID:    requesterID,
		ExecutorID:     cap.AgentID,
		CapabilityID:   cap.ID,
		Input:          req.Input,
		Status:         StatusPending,
		CallbackURL:    req.CallbackURL,
		CallbackSecret: req.CallbackSecret,
		PriceAmount:    price,
		PriceCurrency:  currency,
		DeadlineAt:     req.DeadlineAt,
		Metadata:       req.Metadata,
		MaxRetries:     3,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// 8. Record initial status
	s.repo.RecordStatusHistory(ctx, &TaskStatusHistory{
		TaskID:    task.ID,
		ToStatus:  StatusPending,
		CreatedAt: now,
	})

	// 9. Publish event to notify executor
	s.publishEvent(ctx, "task.created", map[string]any{
		"task_id":       task.ID,
		"capability_id": task.CapabilityID,
		"requester_id":  task.RequesterID,
		"executor_id":   task.ExecutorID,
		"price":         task.PriceAmount,
		"currency":      task.PriceCurrency,
	})

	// 10. Send callback if configured
	s.sendCallback(ctx, task)

	// Fill in names for response
	task.CapabilityName = cap.Name

	return task, nil
}

// AcceptTask is called by the executor to accept a pending task.
func (s *Service) AcceptTask(ctx context.Context, executorID uuid.UUID, taskID uuid.UUID) (*Task, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	if task.ExecutorID != executorID {
		return nil, ErrNotAuthorized
	}

	if task.Status != StatusPending {
		return nil, fmt.Errorf("%w: task must be pending to accept", ErrInvalidStatus)
	}

	// Update status
	if err := s.repo.UpdateTaskStatus(ctx, taskID, StatusAccepted, "", nil); err != nil {
		return nil, err
	}

	oldStatus := task.Status
	task.Status = StatusAccepted

	// Create transaction for payment
	if s.txCreator != nil {
		txID, err := s.txCreator.CreateFromTask(
			ctx,
			task.RequesterID,
			task.ExecutorID,
			&task.ID,
			task.PriceAmount,
			task.PriceCurrency,
		)
		if err == nil {
			s.repo.SetTransactionID(ctx, taskID, txID)
			task.TransactionID = &txID
		}
	}

	// Record history
	s.repo.RecordStatusHistory(ctx, &TaskStatusHistory{
		TaskID:     taskID,
		FromStatus: &oldStatus,
		ToStatus:   StatusAccepted,
		ChangedBy:  &executorID,
		CreatedAt:  time.Now().UTC(),
	})

	s.publishEvent(ctx, "task.accepted", map[string]any{
		"task_id":        taskID,
		"requester_id":   task.RequesterID,
		"executor_id":    task.ExecutorID,
		"transaction_id": task.TransactionID,
	})

	s.sendCallback(ctx, task)

	return task, nil
}

// UpdateTaskProgress updates task with a custom status event.
func (s *Service) UpdateTaskProgress(ctx context.Context, executorID uuid.UUID, taskID uuid.UUID, req *UpdateTaskProgressRequest) (*Task, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	if task.ExecutorID != executorID {
		return nil, ErrNotAuthorized
	}

	// Must be in accepted or in_progress
	if task.Status != StatusAccepted && task.Status != StatusInProgress {
		return nil, fmt.Errorf("%w: task must be accepted or in_progress", ErrInvalidStatus)
	}

	// Validate event against capability's status_events if provided
	if req.Event != "" {
		cap, _ := s.capability.GetCapabilityByID(ctx, task.CapabilityID)
		if cap != nil && len(cap.StatusEvents) > 0 {
			if !s.isValidEvent(cap.StatusEvents, req.Event) {
				return nil, fmt.Errorf("%w: '%s' is not defined in capability status_events", ErrInvalidEvent, req.Event)
			}
		}
	}

	oldStatus := task.Status
	newStatus := StatusInProgress

	// If first progress update, mark started_at
	if task.Status == StatusAccepted {
		now := time.Now().UTC()
		task.StartedAt = &now
	}

	// Update status and event
	if err := s.repo.UpdateTaskStatus(ctx, taskID, newStatus, req.Event, req.EventData); err != nil {
		return nil, err
	}

	task.Status = newStatus
	task.CurrentEvent = req.Event
	task.CurrentEventData = req.EventData

	// Update started_at in database if needed
	if oldStatus == StatusAccepted {
		s.repo.UpdateTask(ctx, task)
	}

	// Record history
	s.repo.RecordStatusHistory(ctx, &TaskStatusHistory{
		TaskID:     taskID,
		FromStatus: &oldStatus,
		ToStatus:   newStatus,
		Event:      req.Event,
		EventData:  req.EventData,
		ChangedBy:  &executorID,
		CreatedAt:  time.Now().UTC(),
	})

	s.publishEvent(ctx, "task.progress", map[string]any{
		"task_id":      taskID,
		"requester_id": task.RequesterID,
		"executor_id":  task.ExecutorID,
		"event":        req.Event,
		"event_data":   req.EventData,
	})

	s.sendCallback(ctx, task)

	return task, nil
}

// DeliverTask marks task as delivered with output.
func (s *Service) DeliverTask(ctx context.Context, executorID uuid.UUID, taskID uuid.UUID, req *DeliverTaskRequest) (*Task, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	if task.ExecutorID != executorID {
		return nil, ErrNotAuthorized
	}

	if task.Status != StatusAccepted && task.Status != StatusInProgress {
		return nil, fmt.Errorf("%w: task must be accepted or in_progress to deliver", ErrInvalidStatus)
	}

	// Validate output against capability's output_schema
	cap, _ := s.capability.GetCapabilityByID(ctx, task.CapabilityID)
	if cap != nil && s.validator != nil && len(cap.OutputSchema) > 0 {
		if err := s.validator.Validate(cap.OutputSchema, req.Output); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrOutputValidation, err)
		}
	}

	oldStatus := task.Status
	task.Output = req.Output
	task.Status = StatusDelivered

	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return nil, err
	}

	// Record history
	s.repo.RecordStatusHistory(ctx, &TaskStatusHistory{
		TaskID:     taskID,
		FromStatus: &oldStatus,
		ToStatus:   StatusDelivered,
		ChangedBy:  &executorID,
		CreatedAt:  time.Now().UTC(),
	})

	s.publishEvent(ctx, "task.delivered", map[string]any{
		"task_id":      taskID,
		"requester_id": task.RequesterID,
		"executor_id":  task.ExecutorID,
	})

	s.sendCallback(ctx, task)

	return task, nil
}

// ConfirmTask confirms delivery and completes the task (called by requester).
func (s *Service) ConfirmTask(ctx context.Context, requesterID uuid.UUID, taskID uuid.UUID) (*Task, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	if task.RequesterID != requesterID {
		return nil, ErrNotAuthorized
	}

	if task.Status != StatusDelivered {
		return nil, fmt.Errorf("%w: task must be delivered to confirm", ErrInvalidStatus)
	}

	oldStatus := task.Status
	now := time.Now().UTC()
	task.Status = StatusCompleted
	task.CompletedAt = &now

	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return nil, err
	}

	// Update capability stats
	if s.capStats != nil {
		go s.capStats.RecordTaskCompletion(context.Background(), task.CapabilityID, true, nil)
	}

	// Record history
	s.repo.RecordStatusHistory(ctx, &TaskStatusHistory{
		TaskID:     taskID,
		FromStatus: &oldStatus,
		ToStatus:   StatusCompleted,
		ChangedBy:  &requesterID,
		CreatedAt:  now,
	})

	s.publishEvent(ctx, "task.completed", map[string]any{
		"task_id":      taskID,
		"requester_id": task.RequesterID,
		"executor_id":  task.ExecutorID,
	})

	s.sendCallback(ctx, task)

	return task, nil
}

// CancelTask cancels a pending or accepted task (called by requester).
func (s *Service) CancelTask(ctx context.Context, requesterID uuid.UUID, taskID uuid.UUID) (*Task, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	if task.RequesterID != requesterID {
		return nil, ErrNotAuthorized
	}

	// Can only cancel pending or accepted tasks
	if task.Status != StatusPending && task.Status != StatusAccepted {
		return nil, fmt.Errorf("%w: can only cancel pending or accepted tasks", ErrInvalidStatus)
	}

	oldStatus := task.Status
	task.Status = StatusCancelled

	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return nil, err
	}

	// Record history
	s.repo.RecordStatusHistory(ctx, &TaskStatusHistory{
		TaskID:     taskID,
		FromStatus: &oldStatus,
		ToStatus:   StatusCancelled,
		ChangedBy:  &requesterID,
		CreatedAt:  time.Now().UTC(),
	})

	s.publishEvent(ctx, "task.cancelled", map[string]any{
		"task_id":      taskID,
		"requester_id": task.RequesterID,
		"executor_id":  task.ExecutorID,
	})

	s.sendCallback(ctx, task)

	return task, nil
}

// FailTask marks a task as failed (called by executor).
func (s *Service) FailTask(ctx context.Context, executorID uuid.UUID, taskID uuid.UUID, req *FailTaskRequest) (*Task, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	if task.ExecutorID != executorID {
		return nil, ErrNotAuthorized
	}

	if task.Status.IsTerminal() {
		return nil, fmt.Errorf("%w: task is already in terminal state", ErrInvalidStatus)
	}

	oldStatus := task.Status
	task.ErrorMessage = req.ErrorMessage

	// Handle retry logic
	if req.Retry && task.RetryCount < task.MaxRetries {
		task.RetryCount++
		task.Status = StatusPending // Reset to pending for retry
	} else {
		task.Status = StatusFailed
	}

	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return nil, err
	}

	// Update capability stats for failure
	if task.Status == StatusFailed && s.capStats != nil {
		go s.capStats.RecordTaskCompletion(context.Background(), task.CapabilityID, false, nil)
	}

	// Record history
	s.repo.RecordStatusHistory(ctx, &TaskStatusHistory{
		TaskID:     taskID,
		FromStatus: &oldStatus,
		ToStatus:   task.Status,
		ChangedBy:  &executorID,
		CreatedAt:  time.Now().UTC(),
	})

	eventType := "task.failed"
	if task.Status == StatusPending {
		eventType = "task.retry"
	}

	s.publishEvent(ctx, eventType, map[string]any{
		"task_id":       taskID,
		"requester_id":  task.RequesterID,
		"executor_id":   task.ExecutorID,
		"error_message": req.ErrorMessage,
		"retry_count":   task.RetryCount,
	})

	s.sendCallback(ctx, task)

	return task, nil
}

// GetTask retrieves a task by ID.
func (s *Service) GetTask(ctx context.Context, id uuid.UUID) (*Task, error) {
	task, err := s.repo.GetTaskByID(ctx, id)
	if err != nil {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

// ListTasks lists tasks with filters.
func (s *Service) ListTasks(ctx context.Context, params ListTasksParams) (*TaskListResult, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	return s.repo.ListTasks(ctx, params)
}

// GetTaskHistory returns the status history for a task.
func (s *Service) GetTaskHistory(ctx context.Context, taskID uuid.UUID) ([]*TaskStatusHistory, error) {
	return s.repo.GetTaskHistory(ctx, taskID)
}

// --- Helper Methods ---

func (s *Service) calculatePrice(cap *CapabilityInfo, input json.RawMessage) float64 {
	if cap.BaseFee != nil {
		return *cap.BaseFee
	}
	return 0
}

func (s *Service) isValidEvent(statusEvents json.RawMessage, event string) bool {
	if len(statusEvents) == 0 {
		return true // No events defined, allow any
	}

	// Parse status events - they can be either strings or objects with "event" field
	var events []json.RawMessage
	if err := json.Unmarshal(statusEvents, &events); err != nil {
		return true // Can't parse, allow
	}

	for _, e := range events {
		// Try as string first
		var eventStr string
		if err := json.Unmarshal(e, &eventStr); err == nil {
			if eventStr == event {
				return true
			}
			continue
		}

		// Try as object with "event" field
		var eventObj struct {
			Event string `json:"event"`
		}
		if err := json.Unmarshal(e, &eventObj); err == nil {
			if eventObj.Event == event {
				return true
			}
		}
	}

	return false
}

func (s *Service) publishEvent(ctx context.Context, eventType string, payload map[string]any) {
	if s.publisher != nil {
		go s.publisher.Publish(context.Background(), eventType, payload)
	}
}

func (s *Service) sendCallback(ctx context.Context, task *Task) {
	if task.CallbackURL == "" || s.callback == nil {
		return
	}

	callback := TaskCallback{
		TaskID:        task.ID,
		CapabilityID:  task.CapabilityID,
		Status:        task.Status,
		Event:         task.CurrentEvent,
		EventData:     task.CurrentEventData,
		Output:        task.Output,
		Error:         task.ErrorMessage,
		TransactionID: task.TransactionID,
		Timestamp:     time.Now().UTC(),
	}

	// Deliver asynchronously
	go s.callback.DeliverCallback(context.Background(), task.CallbackURL, task.CallbackSecret, callback)
}
