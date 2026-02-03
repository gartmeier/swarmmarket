package storage

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

// AllowedImageTypes defines the allowed MIME types for uploads.
var AllowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// Thumbnail sizes
const (
	ThumbnailWidth  = 400
	ThumbnailHeight = 400
	AvatarSize      = 256
)

// EntityType represents the type of entity the image belongs to.
type EntityType string

const (
	EntityTypeListing EntityType = "listings"
	EntityTypeRequest EntityType = "requests"
	EntityTypeAuction EntityType = "auctions"
	EntityTypeAvatar  EntityType = "avatars"
)

// Config holds storage configuration.
type Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	PublicURL       string
	MaxFileSizeMB   int
}

// Service handles image uploads to Cloudflare R2.
type Service struct {
	client       *s3.Client
	bucketName   string
	publicURL    string
	maxFileSizeB int64
}

// NewService creates a new storage service.
func NewService(cfg Config) (*Service, error) {
	if cfg.AccountID == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return nil, fmt.Errorf("R2 credentials not configured")
	}

	// Create R2 endpoint URL
	r2Endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)

	// Create custom resolver for R2
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: r2Endpoint,
		}, nil
	})

	// Load AWS config with R2 credentials
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(awsCfg)

	maxSize := int64(cfg.MaxFileSizeMB) * 1024 * 1024
	if maxSize == 0 {
		maxSize = 10 * 1024 * 1024 // Default 10MB
	}

	return &Service{
		client:       client,
		bucketName:   cfg.BucketName,
		publicURL:    cfg.PublicURL,
		maxFileSizeB: maxSize,
	}, nil
}

// UploadResult contains URLs for the uploaded image and its thumbnail.
type UploadResult struct {
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// UploadImage uploads an image and returns the public URL (no thumbnail).
func (s *Service) UploadImage(ctx context.Context, entityType EntityType, entityID uuid.UUID, data []byte) (string, error) {
	result, err := s.uploadImageInternal(ctx, entityType, entityID, data, false)
	if err != nil {
		return "", err
	}
	return result.URL, nil
}

// UploadImageWithThumbnail uploads an image with automatic thumbnail generation.
func (s *Service) UploadImageWithThumbnail(ctx context.Context, entityType EntityType, entityID uuid.UUID, data []byte) (*UploadResult, error) {
	return s.uploadImageInternal(ctx, entityType, entityID, data, true)
}

// uploadImageInternal handles the actual upload logic.
func (s *Service) uploadImageInternal(ctx context.Context, entityType EntityType, entityID uuid.UUID, data []byte, generateThumbnail bool) (*UploadResult, error) {
	// Validate file size
	if int64(len(data)) > s.maxFileSizeB {
		return nil, fmt.Errorf("file size exceeds maximum allowed (%d MB)", s.maxFileSizeB/(1024*1024))
	}

	// Detect content type
	contentType := http.DetectContentType(data)
	ext, ok := AllowedImageTypes[contentType]
	if !ok {
		return nil, fmt.Errorf("unsupported image type: %s (allowed: JPEG, PNG, GIF, WebP)", contentType)
	}

	// Generate unique base filename
	baseID := uuid.New().String()
	filename := fmt.Sprintf("%s/%s/%s%s",
		entityType,
		entityID.String(),
		baseID,
		ext,
	)

	// Upload original to R2
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(s.bucketName),
		Key:          aws.String(filename),
		Body:         bytes.NewReader(data),
		ContentType:  aws.String(contentType),
		CacheControl: aws.String("public, max-age=31536000"), // 1 year cache
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	result := &UploadResult{
		URL: s.GetPublicURL(filename),
	}

	// Generate and upload thumbnail if requested
	if generateThumbnail {
		thumbData, thumbContentType, err := s.generateThumbnail(data, ThumbnailWidth, ThumbnailHeight)
		if err == nil && thumbData != nil {
			thumbExt := AllowedImageTypes[thumbContentType]
			thumbFilename := fmt.Sprintf("%s/%s/%s_thumb%s",
				entityType,
				entityID.String(),
				baseID,
				thumbExt,
			)

			_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
				Bucket:       aws.String(s.bucketName),
				Key:          aws.String(thumbFilename),
				Body:         bytes.NewReader(thumbData),
				ContentType:  aws.String(thumbContentType),
				CacheControl: aws.String("public, max-age=31536000"),
			})
			if err == nil {
				result.ThumbnailURL = s.GetPublicURL(thumbFilename)
			}
		}
	}

	return result, nil
}

// UploadAvatar uploads an agent avatar image (resized to standard size).
func (s *Service) UploadAvatar(ctx context.Context, agentID uuid.UUID, data []byte) (string, error) {
	// Validate file size
	if int64(len(data)) > s.maxFileSizeB {
		return "", fmt.Errorf("file size exceeds maximum allowed (%d MB)", s.maxFileSizeB/(1024*1024))
	}

	// Detect content type
	contentType := http.DetectContentType(data)
	if !ValidateImageType(contentType) {
		return "", fmt.Errorf("unsupported image type: %s (allowed: JPEG, PNG, GIF, WebP)", contentType)
	}

	// Resize avatar to standard size
	resizedData, resizedContentType, err := s.resizeAvatar(data, AvatarSize)
	if err != nil {
		return "", fmt.Errorf("failed to process avatar: %w", err)
	}

	ext := AllowedImageTypes[resizedContentType]
	filename := fmt.Sprintf("%s/%s/avatar%s",
		EntityTypeAvatar,
		agentID.String(),
		ext,
	)

	// Upload to R2
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(s.bucketName),
		Key:          aws.String(filename),
		Body:         bytes.NewReader(resizedData),
		ContentType:  aws.String(resizedContentType),
		CacheControl: aws.String("public, max-age=86400"), // 1 day cache for avatars
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload avatar: %w", err)
	}

	return s.GetPublicURL(filename), nil
}

// generateThumbnail creates a thumbnail from image data.
func (s *Service) generateThumbnail(data []byte, maxWidth, maxHeight int) ([]byte, string, error) {
	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize using imaging library (maintains aspect ratio, fits within bounds)
	thumbnail := imaging.Fit(img, maxWidth, maxHeight, imaging.Lanczos)

	// Encode as JPEG for smaller size
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 85}); err != nil {
		return nil, "", fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return buf.Bytes(), "image/jpeg", nil
}

// resizeAvatar resizes an avatar to a square image.
func (s *Service) resizeAvatar(data []byte, size int) ([]byte, string, error) {
	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Crop to square and resize (center crop)
	avatar := imaging.Fill(img, size, size, imaging.Center, imaging.Lanczos)

	// Encode as PNG for avatars (supports transparency)
	var buf bytes.Buffer
	if err := png.Encode(&buf, avatar); err != nil {
		return nil, "", fmt.Errorf("failed to encode avatar: %w", err)
	}

	return buf.Bytes(), "image/png", nil
}

// UploadImageFromReader uploads an image from an io.Reader.
func (s *Service) UploadImageFromReader(ctx context.Context, entityType EntityType, entityID uuid.UUID, reader io.Reader, contentType string) (string, error) {
	// Read all data (needed for size validation)
	data, err := io.ReadAll(io.LimitReader(reader, s.maxFileSizeB+1))
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}

	return s.UploadImage(ctx, entityType, entityID, data)
}

// DeleteImage deletes an image by its key.
func (s *Service) DeleteImage(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}
	return nil
}

// DeleteEntityImages deletes all images for an entity.
func (s *Service) DeleteEntityImages(ctx context.Context, entityType EntityType, entityID uuid.UUID) error {
	prefix := fmt.Sprintf("%s/%s/", entityType, entityID.String())

	// List all objects with the prefix
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list images: %w", err)
		}

		// Delete each object
		for _, obj := range page.Contents {
			if err := s.DeleteImage(ctx, *obj.Key); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetPublicURL returns the public URL for a file key.
func (s *Service) GetPublicURL(key string) string {
	if s.publicURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.publicURL, "/"), key)
	}
	// Fallback to R2 public access URL (requires public bucket)
	return fmt.Sprintf("https://%s.r2.dev/%s", s.bucketName, key)
}

// ExtractKeyFromURL extracts the object key from a public URL.
func (s *Service) ExtractKeyFromURL(url string) string {
	// Remove public URL prefix
	if s.publicURL != "" {
		return strings.TrimPrefix(url, strings.TrimSuffix(s.publicURL, "/")+"/")
	}
	// Remove R2.dev prefix
	prefix := fmt.Sprintf("https://%s.r2.dev/", s.bucketName)
	return strings.TrimPrefix(url, prefix)
}

// Image represents an uploaded image with metadata.
type Image struct {
	ID           uuid.UUID  `json:"id"`
	EntityType   EntityType `json:"entity_type"`
	EntityID     uuid.UUID  `json:"entity_id"`
	URL          string     `json:"url"`
	ThumbnailURL string     `json:"thumbnail_url,omitempty"`
	Filename     string     `json:"filename"`
	Size         int64      `json:"size"`
	MimeType     string     `json:"mime_type"`
	CreatedAt    time.Time  `json:"created_at"`
}

// ValidateImageType checks if the content type is allowed.
func ValidateImageType(contentType string) bool {
	_, ok := AllowedImageTypes[contentType]
	return ok
}

// GetExtensionForType returns the file extension for a content type.
func GetExtensionForType(contentType string) string {
	ext, ok := AllowedImageTypes[contentType]
	if !ok {
		return ".bin"
	}
	return ext
}

// GetContentTypeFromExtension returns the content type for a file extension.
func GetContentTypeFromExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	for mimeType, e := range AllowedImageTypes {
		if e == ext {
			return mimeType
		}
	}
	return "application/octet-stream"
}
