package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles image metadata persistence.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new image repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create stores image metadata in the database.
func (r *Repository) Create(ctx context.Context, img *Image) error {
	query := `
		INSERT INTO entity_images (id, entity_type, entity_id, url, thumbnail_url, filename, size_bytes, mime_type, position, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8,
			COALESCE((SELECT MAX(position) + 1 FROM entity_images WHERE entity_type = $2 AND entity_id = $3), 0),
			$9)
		RETURNING position
	`
	img.ID = uuid.New()
	if img.CreatedAt.IsZero() {
		img.CreatedAt = time.Now().UTC()
	}

	var position int
	err := r.db.QueryRow(ctx, query,
		img.ID,
		img.EntityType,
		img.EntityID,
		img.URL,
		img.ThumbnailURL,
		img.Filename,
		img.Size,
		img.MimeType,
		img.CreatedAt,
	).Scan(&position)

	return err
}

// GetByEntity returns all images for an entity.
func (r *Repository) GetByEntity(ctx context.Context, entityType EntityType, entityID uuid.UUID) ([]Image, error) {
	query := `
		SELECT id, entity_type, entity_id, url, COALESCE(thumbnail_url, ''), filename, size_bytes, mime_type, created_at
		FROM entity_images
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY position ASC
	`

	rows, err := r.db.Query(ctx, query, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get images: %w", err)
	}
	defer rows.Close()

	var images []Image
	for rows.Next() {
		var img Image
		if err := rows.Scan(
			&img.ID,
			&img.EntityType,
			&img.EntityID,
			&img.URL,
			&img.ThumbnailURL,
			&img.Filename,
			&img.Size,
			&img.MimeType,
			&img.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}
		images = append(images, img)
	}

	return images, nil
}

// GetFirstByEntity returns the first image for an entity (for thumbnails).
func (r *Repository) GetFirstByEntity(ctx context.Context, entityType EntityType, entityID uuid.UUID) (*Image, error) {
	query := `
		SELECT id, entity_type, entity_id, url, COALESCE(thumbnail_url, ''), filename, size_bytes, mime_type, created_at
		FROM entity_images
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY position ASC
		LIMIT 1
	`

	var img Image
	err := r.db.QueryRow(ctx, query, entityType, entityID).Scan(
		&img.ID,
		&img.EntityType,
		&img.EntityID,
		&img.URL,
		&img.ThumbnailURL,
		&img.Filename,
		&img.Size,
		&img.MimeType,
		&img.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &img, nil
}

// Delete removes an image by ID.
func (r *Repository) Delete(ctx context.Context, imageID uuid.UUID) (*Image, error) {
	query := `
		DELETE FROM entity_images
		WHERE id = $1
		RETURNING id, entity_type, entity_id, url, COALESCE(thumbnail_url, ''), filename, size_bytes, mime_type, created_at
	`

	var img Image
	err := r.db.QueryRow(ctx, query, imageID).Scan(
		&img.ID,
		&img.EntityType,
		&img.EntityID,
		&img.URL,
		&img.ThumbnailURL,
		&img.Filename,
		&img.Size,
		&img.MimeType,
		&img.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to delete image: %w", err)
	}

	return &img, nil
}

// DeleteByEntity removes all images for an entity.
func (r *Repository) DeleteByEntity(ctx context.Context, entityType EntityType, entityID uuid.UUID) ([]Image, error) {
	query := `
		DELETE FROM entity_images
		WHERE entity_type = $1 AND entity_id = $2
		RETURNING id, entity_type, entity_id, url, COALESCE(thumbnail_url, ''), filename, size_bytes, mime_type, created_at
	`

	rows, err := r.db.Query(ctx, query, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete images: %w", err)
	}
	defer rows.Close()

	var images []Image
	for rows.Next() {
		var img Image
		if err := rows.Scan(
			&img.ID,
			&img.EntityType,
			&img.EntityID,
			&img.URL,
			&img.ThumbnailURL,
			&img.Filename,
			&img.Size,
			&img.MimeType,
			&img.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan deleted image: %w", err)
		}
		images = append(images, img)
	}

	return images, nil
}

// CountByEntity returns the number of images for an entity.
func (r *Repository) CountByEntity(ctx context.Context, entityType EntityType, entityID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM entity_images WHERE entity_type = $1 AND entity_id = $2`

	var count int
	err := r.db.QueryRow(ctx, query, entityType, entityID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count images: %w", err)
	}

	return count, nil
}

// EntityImageURLs holds the URL and thumbnail URL for an entity's first image.
type EntityImageURLs struct {
	URL          string
	ThumbnailURL string
}

// GetImageURLsForEntities returns the first image URLs for multiple entities (for list views).
func (r *Repository) GetImageURLsForEntities(ctx context.Context, entityType EntityType, entityIDs []uuid.UUID) (map[uuid.UUID]EntityImageURLs, error) {
	if len(entityIDs) == 0 {
		return make(map[uuid.UUID]EntityImageURLs), nil
	}

	query := `
		SELECT DISTINCT ON (entity_id) entity_id, url, COALESCE(thumbnail_url, '')
		FROM entity_images
		WHERE entity_type = $1 AND entity_id = ANY($2)
		ORDER BY entity_id, position ASC
	`

	rows, err := r.db.Query(ctx, query, entityType, entityIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get image URLs: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]EntityImageURLs)
	for rows.Next() {
		var entityID uuid.UUID
		var urls EntityImageURLs
		if err := rows.Scan(&entityID, &urls.URL, &urls.ThumbnailURL); err != nil {
			return nil, fmt.Errorf("failed to scan image URL: %w", err)
		}
		result[entityID] = urls
	}

	return result, nil
}
