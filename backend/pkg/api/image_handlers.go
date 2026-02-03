package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/digi604/swarmmarket/backend/internal/common"
	"github.com/digi604/swarmmarket/backend/internal/storage"
	"github.com/digi604/swarmmarket/backend/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	maxUploadSize      = 10 << 20 // 10 MB
	maxImagesPerEntity = 10
)

// ImageHandler handles image upload/delete requests.
type ImageHandler struct {
	storageService   *storage.Service
	imageRepo        *storage.Repository
	ownershipChecker OwnershipChecker
	agentUpdater     AgentAvatarUpdater
}

// OwnershipChecker verifies entity ownership.
type OwnershipChecker interface {
	IsListingOwner(ctx context.Context, listingID, agentID uuid.UUID) (bool, error)
	IsRequestOwner(ctx context.Context, requestID, agentID uuid.UUID) (bool, error)
	IsAuctionOwner(ctx context.Context, auctionID, agentID uuid.UUID) (bool, error)
}

// AgentAvatarUpdater updates an agent's avatar URL.
type AgentAvatarUpdater interface {
	UpdateAvatarURL(ctx context.Context, agentID uuid.UUID, avatarURL string) error
}

// NewImageHandler creates a new image handler.
func NewImageHandler(storageService *storage.Service, imageRepo *storage.Repository, ownershipChecker OwnershipChecker) *ImageHandler {
	return &ImageHandler{
		storageService:   storageService,
		imageRepo:        imageRepo,
		ownershipChecker: ownershipChecker,
	}
}

// SetAgentUpdater sets the agent avatar updater (to avoid circular dependency).
func (h *ImageHandler) SetAgentUpdater(updater AgentAvatarUpdater) {
	h.agentUpdater = updater
}

// UploadListingImage handles image uploads for listings.
func (h *ImageHandler) UploadListingImage(w http.ResponseWriter, r *http.Request) {
	h.uploadImage(w, r, storage.EntityTypeListing, func(entityID, agentID uuid.UUID) (bool, error) {
		return h.ownershipChecker.IsListingOwner(r.Context(), entityID, agentID)
	})
}

// UploadRequestImage handles image uploads for requests.
func (h *ImageHandler) UploadRequestImage(w http.ResponseWriter, r *http.Request) {
	h.uploadImage(w, r, storage.EntityTypeRequest, func(entityID, agentID uuid.UUID) (bool, error) {
		return h.ownershipChecker.IsRequestOwner(r.Context(), entityID, agentID)
	})
}

// UploadAuctionImage handles image uploads for auctions.
func (h *ImageHandler) UploadAuctionImage(w http.ResponseWriter, r *http.Request) {
	h.uploadImage(w, r, storage.EntityTypeAuction, func(entityID, agentID uuid.UUID) (bool, error) {
		return h.ownershipChecker.IsAuctionOwner(r.Context(), entityID, agentID)
	})
}

// UploadAvatar handles avatar image uploads for agents.
func (h *ImageHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authenticated agent
	agent := middleware.GetAgent(ctx)
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("authentication required"))
		return
	}

	// Limit upload size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Parse multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("file too large or invalid form data"))
		return
	}

	// Get the file
	file, _, err := r.FormFile("image")
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("image file required"))
		return
	}
	defer file.Close()

	// Read file data
	data, err := io.ReadAll(file)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("failed to read file"))
		return
	}

	// Detect content type
	contentType := http.DetectContentType(data)
	if !storage.ValidateImageType(contentType) {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("unsupported image type (allowed: JPEG, PNG, GIF, WebP)"))
		return
	}

	// Upload avatar to R2 (automatically resized)
	avatarURL, err := h.storageService.UploadAvatar(ctx, agent.ID, data)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer(fmt.Sprintf("failed to upload avatar: %v", err)))
		return
	}

	// Update agent's avatar URL in database
	if h.agentUpdater != nil {
		if err := h.agentUpdater.UpdateAvatarURL(ctx, agent.ID, avatarURL); err != nil {
			common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to update agent avatar"))
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"avatar_url": avatarURL,
	})
}

// uploadImage is the common image upload logic with automatic thumbnail generation.
func (h *ImageHandler) uploadImage(w http.ResponseWriter, r *http.Request, entityType storage.EntityType, checkOwnership func(uuid.UUID, uuid.UUID) (bool, error)) {
	ctx := r.Context()

	// Get authenticated agent
	agent := middleware.GetAgent(ctx)
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("authentication required"))
		return
	}

	// Parse entity ID from URL
	idStr := chi.URLParam(r, "id")
	entityID, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid entity ID"))
		return
	}

	// Check ownership
	isOwner, err := checkOwnership(entityID, agent.ID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to verify ownership"))
		return
	}
	if !isOwner {
		common.WriteError(w, http.StatusForbidden, common.ErrForbidden("you can only upload images to your own entities"))
		return
	}

	// Check image count limit
	count, err := h.imageRepo.CountByEntity(ctx, entityType, entityID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to count images"))
		return
	}
	if count >= maxImagesPerEntity {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest(fmt.Sprintf("maximum %d images allowed per entity", maxImagesPerEntity)))
		return
	}

	// Limit upload size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Parse multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("file too large or invalid form data"))
		return
	}

	// Get the file
	file, header, err := r.FormFile("image")
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("image file required"))
		return
	}
	defer file.Close()

	// Read file data
	data, err := io.ReadAll(file)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("failed to read file"))
		return
	}

	// Detect content type
	contentType := http.DetectContentType(data)
	if !storage.ValidateImageType(contentType) {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("unsupported image type (allowed: JPEG, PNG, GIF, WebP)"))
		return
	}

	// Upload to R2 with automatic thumbnail generation
	result, err := h.storageService.UploadImageWithThumbnail(ctx, entityType, entityID, data)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer(fmt.Sprintf("failed to upload image: %v", err)))
		return
	}

	// Save metadata to database
	img := &storage.Image{
		EntityType:   entityType,
		EntityID:     entityID,
		URL:          result.URL,
		ThumbnailURL: result.ThumbnailURL,
		Filename:     header.Filename,
		Size:         int64(len(data)),
		MimeType:     contentType,
	}
	if err := h.imageRepo.Create(ctx, img); err != nil {
		// Try to delete the uploaded files on DB error
		_ = h.storageService.DeleteImage(ctx, h.storageService.ExtractKeyFromURL(result.URL))
		if result.ThumbnailURL != "" {
			_ = h.storageService.DeleteImage(ctx, h.storageService.ExtractKeyFromURL(result.ThumbnailURL))
		}
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to save image metadata"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(img)
}

// GetListingImages returns images for a listing.
func (h *ImageHandler) GetListingImages(w http.ResponseWriter, r *http.Request) {
	h.getImages(w, r, storage.EntityTypeListing)
}

// GetRequestImages returns images for a request.
func (h *ImageHandler) GetRequestImages(w http.ResponseWriter, r *http.Request) {
	h.getImages(w, r, storage.EntityTypeRequest)
}

// GetAuctionImages returns images for an auction.
func (h *ImageHandler) GetAuctionImages(w http.ResponseWriter, r *http.Request) {
	h.getImages(w, r, storage.EntityTypeAuction)
}

// getImages is the common get images logic.
func (h *ImageHandler) getImages(w http.ResponseWriter, r *http.Request, entityType storage.EntityType) {
	ctx := r.Context()

	// Parse entity ID from URL
	idStr := chi.URLParam(r, "id")
	entityID, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid entity ID"))
		return
	}

	images, err := h.imageRepo.GetByEntity(ctx, entityType, entityID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to get images"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"images": images,
		"count":  len(images),
	})
}

// DeleteListingImage deletes an image from a listing.
func (h *ImageHandler) DeleteListingImage(w http.ResponseWriter, r *http.Request) {
	h.deleteImage(w, r, storage.EntityTypeListing, func(entityID, agentID uuid.UUID) (bool, error) {
		return h.ownershipChecker.IsListingOwner(r.Context(), entityID, agentID)
	})
}

// DeleteRequestImage deletes an image from a request.
func (h *ImageHandler) DeleteRequestImage(w http.ResponseWriter, r *http.Request) {
	h.deleteImage(w, r, storage.EntityTypeRequest, func(entityID, agentID uuid.UUID) (bool, error) {
		return h.ownershipChecker.IsRequestOwner(r.Context(), entityID, agentID)
	})
}

// DeleteAuctionImage deletes an image from an auction.
func (h *ImageHandler) DeleteAuctionImage(w http.ResponseWriter, r *http.Request) {
	h.deleteImage(w, r, storage.EntityTypeAuction, func(entityID, agentID uuid.UUID) (bool, error) {
		return h.ownershipChecker.IsAuctionOwner(r.Context(), entityID, agentID)
	})
}

// deleteImage is the common delete image logic.
func (h *ImageHandler) deleteImage(w http.ResponseWriter, r *http.Request, entityType storage.EntityType, checkOwnership func(uuid.UUID, uuid.UUID) (bool, error)) {
	ctx := r.Context()

	// Get authenticated agent
	agent := middleware.GetAgent(ctx)
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("authentication required"))
		return
	}

	// Parse entity ID and image ID from URL
	idStr := chi.URLParam(r, "id")
	entityID, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid entity ID"))
		return
	}

	imageIDStr := chi.URLParam(r, "imageId")
	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid image ID"))
		return
	}

	// Check ownership
	isOwner, err := checkOwnership(entityID, agent.ID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to verify ownership"))
		return
	}
	if !isOwner {
		common.WriteError(w, http.StatusForbidden, common.ErrForbidden("you can only delete images from your own entities"))
		return
	}

	// Delete from database and get URLs for R2 deletion
	img, err := h.imageRepo.Delete(ctx, imageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("image not found"))
			return
		}
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to delete image"))
		return
	}

	// Delete original from R2 (best effort)
	key := h.storageService.ExtractKeyFromURL(img.URL)
	_ = h.storageService.DeleteImage(ctx, key)

	// Delete thumbnail from R2 if exists (best effort)
	if img.ThumbnailURL != "" {
		thumbKey := h.storageService.ExtractKeyFromURL(img.ThumbnailURL)
		_ = h.storageService.DeleteImage(ctx, thumbKey)
	}

	w.WriteHeader(http.StatusNoContent)
}
