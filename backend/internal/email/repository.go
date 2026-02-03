package email

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles database operations for email queue.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new email repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// QueueEmail adds an email to the queue.
func (r *Repository) QueueEmail(ctx context.Context, email *QueuedEmail) error {
	payloadJSON, err := json.Marshal(email.Payload)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO email_queue (id, recipient_email, recipient_agent_id, template, payload, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = r.pool.Exec(ctx, query,
		email.ID, email.RecipientEmail, email.RecipientAgentID,
		email.Template, payloadJSON, email.Status, email.CreatedAt,
	)
	return err
}

// GetPendingEmails retrieves emails ready to be sent.
func (r *Repository) GetPendingEmails(ctx context.Context, limit int) ([]*QueuedEmail, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT id, recipient_email, recipient_agent_id, template, payload, status, attempts, last_attempt_at, sent_at, error, created_at
		FROM email_queue
		WHERE status = 'pending' AND attempts < 5
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []*QueuedEmail
	for rows.Next() {
		var email QueuedEmail
		var payloadJSON []byte
		if err := rows.Scan(
			&email.ID, &email.RecipientEmail, &email.RecipientAgentID,
			&email.Template, &payloadJSON, &email.Status, &email.Attempts,
			&email.LastAttemptAt, &email.SentAt, &email.Error, &email.CreatedAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(payloadJSON, &email.Payload); err != nil {
			email.Payload = map[string]any{}
		}
		emails = append(emails, &email)
	}

	return emails, nil
}

// MarkEmailSent marks an email as successfully sent.
func (r *Repository) MarkEmailSent(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE email_queue
		SET status = 'sent', sent_at = $1, attempts = attempts + 1, last_attempt_at = $1
		WHERE id = $2
	`
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), id)
	return err
}

// MarkEmailFailed marks an email as failed with error message.
func (r *Repository) MarkEmailFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	now := time.Now().UTC()
	query := `
		UPDATE email_queue
		SET attempts = attempts + 1, last_attempt_at = $1, error = $2,
			status = CASE WHEN attempts >= 4 THEN 'failed' ELSE 'pending' END
		WHERE id = $3
	`
	_, err := r.pool.Exec(ctx, query, now, errMsg, id)
	return err
}

// GetRecentEmailToRecipient checks if an email was sent recently to avoid spam.
func (r *Repository) GetRecentEmailToRecipient(ctx context.Context, recipientAgentID uuid.UUID, template EmailTemplate, cooldownMinutes int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM email_queue
			WHERE recipient_agent_id = $1
			AND template = $2
			AND status = 'sent'
			AND sent_at > NOW() - INTERVAL '1 minute' * $3
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, query, recipientAgentID, template, cooldownMinutes).Scan(&exists)
	return exists, err
}

// CleanupOldEmails removes old processed emails.
func (r *Repository) CleanupOldEmails(ctx context.Context, olderThanDays int) error {
	query := `
		DELETE FROM email_queue
		WHERE (status = 'sent' OR status = 'failed')
		AND created_at < NOW() - INTERVAL '1 day' * $1
	`
	_, err := r.pool.Exec(ctx, query, olderThanDays)
	return err
}
