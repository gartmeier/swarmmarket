package notification

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrWebhookNotFound = errors.New("webhook not found")
)

// Repository handles webhook persistence.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new notification repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// CreateWebhook creates a new webhook for an agent.
func (r *Repository) CreateWebhook(ctx context.Context, agentID uuid.UUID, url string, eventTypes []string) (*Webhook, error) {
	// Generate secret for HMAC signing
	secret, err := generateSecret(32)
	if err != nil {
		return nil, err
	}

	webhook := &Webhook{
		ID:         uuid.New(),
		AgentID:    agentID,
		URL:        url,
		Secret:     secret,
		EventTypes: make([]EventType, len(eventTypes)),
		IsActive:   true,
		CreatedAt:  time.Now().UTC(),
	}

	for i, et := range eventTypes {
		webhook.EventTypes[i] = EventType(et)
	}

	query := `
		INSERT INTO webhooks (id, agent_id, url, secret, events, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.pool.Exec(ctx, query,
		webhook.ID,
		webhook.AgentID,
		webhook.URL,
		webhook.Secret,
		eventTypes,
		webhook.IsActive,
		webhook.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return webhook, nil
}

// GetWebhookByID retrieves a webhook by ID.
func (r *Repository) GetWebhookByID(ctx context.Context, id uuid.UUID) (*Webhook, error) {
	query := `
		SELECT id, agent_id, url, secret, events, is_active, last_triggered_at, failure_count, created_at
		FROM webhooks
		WHERE id = $1
	`

	var webhook Webhook
	var eventTypes []string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&webhook.ID,
		&webhook.AgentID,
		&webhook.URL,
		&webhook.Secret,
		&eventTypes,
		&webhook.IsActive,
		&webhook.LastTriggeredAt,
		&webhook.FailureCount,
		&webhook.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWebhookNotFound
		}
		return nil, err
	}

	webhook.EventTypes = make([]EventType, len(eventTypes))
	for i, et := range eventTypes {
		webhook.EventTypes[i] = EventType(et)
	}

	return &webhook, nil
}

// GetWebhooksByAgentID retrieves all webhooks for an agent.
func (r *Repository) GetWebhooksByAgentID(ctx context.Context, agentID uuid.UUID) ([]*Webhook, error) {
	query := `
		SELECT id, agent_id, url, secret, events, is_active, last_triggered_at, failure_count, created_at
		FROM webhooks
		WHERE agent_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []*Webhook
	for rows.Next() {
		var webhook Webhook
		var eventTypes []string
		if err := rows.Scan(
			&webhook.ID,
			&webhook.AgentID,
			&webhook.URL,
			&webhook.Secret,
			&eventTypes,
			&webhook.IsActive,
			&webhook.LastTriggeredAt,
			&webhook.FailureCount,
			&webhook.CreatedAt,
		); err != nil {
			return nil, err
		}

		webhook.EventTypes = make([]EventType, len(eventTypes))
		for i, et := range eventTypes {
			webhook.EventTypes[i] = EventType(et)
		}

		webhooks = append(webhooks, &webhook)
	}

	return webhooks, nil
}

// GetActiveWebhooksForEvent retrieves all active webhooks subscribed to an event type.
func (r *Repository) GetActiveWebhooksForEvent(ctx context.Context, eventType string) ([]*Webhook, error) {
	query := `
		SELECT id, agent_id, url, secret, events, is_active, last_triggered_at, failure_count, created_at
		FROM webhooks
		WHERE is_active = true AND $1 = ANY(events)
	`

	rows, err := r.pool.Query(ctx, query, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []*Webhook
	for rows.Next() {
		var webhook Webhook
		var eventTypes []string
		if err := rows.Scan(
			&webhook.ID,
			&webhook.AgentID,
			&webhook.URL,
			&webhook.Secret,
			&eventTypes,
			&webhook.IsActive,
			&webhook.LastTriggeredAt,
			&webhook.FailureCount,
			&webhook.CreatedAt,
		); err != nil {
			return nil, err
		}

		webhook.EventTypes = make([]EventType, len(eventTypes))
		for i, et := range eventTypes {
			webhook.EventTypes[i] = EventType(et)
		}

		webhooks = append(webhooks, &webhook)
	}

	return webhooks, nil
}

// DeleteWebhook deletes a webhook.
func (r *Repository) DeleteWebhook(ctx context.Context, id, agentID uuid.UUID) error {
	query := `DELETE FROM webhooks WHERE id = $1 AND agent_id = $2`

	result, err := r.pool.Exec(ctx, query, id, agentID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrWebhookNotFound
	}

	return nil
}

// UpdateWebhookStatus updates the active status of a webhook.
func (r *Repository) UpdateWebhookStatus(ctx context.Context, id uuid.UUID, isActive bool) error {
	query := `UPDATE webhooks SET is_active = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, isActive)
	return err
}

// RecordWebhookFailure increments the failure count for a webhook.
func (r *Repository) RecordWebhookFailure(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE webhooks
		SET failure_count = failure_count + 1, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// RecordWebhookSuccess resets failure count and updates last triggered time.
func (r *Repository) RecordWebhookSuccess(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE webhooks
		SET failure_count = 0, last_triggered_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func generateSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// InsertEvent persists an event to the database for activity logging.
func (r *Repository) InsertEvent(ctx context.Context, event *Event) error {
	query := `
		INSERT INTO events (id, event_type, agent_id, resource_type, resource_id, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	// Extract resource info from payload
	var resourceType *string
	var resourceID *uuid.UUID

	if rt, ok := event.Payload["resource_type"].(string); ok {
		resourceType = &rt
	}
	if rid, ok := event.Payload["resource_id"].(string); ok {
		if parsed, err := uuid.Parse(rid); err == nil {
			resourceID = &parsed
		}
	}

	_, err := r.pool.Exec(ctx, query,
		event.ID,
		string(event.Type),
		event.AgentID,
		resourceType,
		resourceID,
		event.Payload,
		event.CreatedAt,
	)
	return err
}

// ActivityEvent represents an event for the activity feed.
type ActivityEvent struct {
	ID           uuid.UUID      `json:"id"`
	EventType    string         `json:"event_type"`
	AgentID      uuid.UUID      `json:"agent_id"`
	ResourceType *string        `json:"resource_type,omitempty"`
	ResourceID   *uuid.UUID     `json:"resource_id,omitempty"`
	Payload      map[string]any `json:"payload"`
	CreatedAt    time.Time      `json:"created_at"`
}

// GetAgentActivity retrieves activity events for an agent.
func (r *Repository) GetAgentActivity(ctx context.Context, agentID uuid.UUID, limit, offset int) ([]*ActivityEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT id, event_type, agent_id, resource_type, resource_id, payload, created_at
		FROM events
		WHERE agent_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, agentID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*ActivityEvent
	for rows.Next() {
		var event ActivityEvent
		if err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.AgentID,
			&event.ResourceType,
			&event.ResourceID,
			&event.Payload,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, &event)
	}

	if events == nil {
		events = []*ActivityEvent{}
	}

	return events, nil
}

// GetAgentActivityCount returns the total count of activity events for an agent.
func (r *Repository) GetAgentActivityCount(ctx context.Context, agentID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM events WHERE agent_id = $1`
	var count int
	err := r.pool.QueryRow(ctx, query, agentID).Scan(&count)
	return count, err
}
