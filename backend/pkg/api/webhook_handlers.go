package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/digi604/swarmmarket/backend/internal/common"
	"github.com/digi604/swarmmarket/backend/internal/notification"
	"github.com/digi604/swarmmarket/backend/pkg/middleware"
)

// WebhookHandler handles webhook HTTP requests.
type WebhookHandler struct {
	repo *notification.Repository
}

// NewWebhookHandler creates a new webhook handler.
func NewWebhookHandler(repo *notification.Repository) *WebhookHandler {
	return &WebhookHandler{repo: repo}
}

// CreateWebhookRequest is the request body for creating a webhook.
type CreateWebhookRequest struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

// WebhookResponse is the response for webhook endpoints.
type WebhookResponse struct {
	ID         uuid.UUID                  `json:"id"`
	URL        string                     `json:"url"`
	Secret     string                     `json:"secret,omitempty"` // Only on create
	Events     []notification.EventType   `json:"events"`
	IsActive   bool                       `json:"is_active"`
	CreatedAt  string                     `json:"created_at"`
}

// CreateWebhook handles POST /webhooks - create a new webhook.
func (h *WebhookHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	var req CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	if req.URL == "" {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("url is required"))
		return
	}

	if len(req.Events) == 0 {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("at least one event type is required"))
		return
	}

	// Validate event types
	validEvents := map[string]bool{
		"request.created":    true,
		"offer.received":     true,
		"offer.accepted":     true,
		"offer.rejected":     true,
		"listing.created":    true,
		"listing.updated":    true,
		"auction.started":    true,
		"bid.placed":         true,
		"bid.outbid":         true,
		"auction.ending_soon": true,
		"auction.ended":      true,
		"order.created":      true,
		"escrow.funded":      true,
		"delivery.confirmed": true,
		"payment.released":   true,
		"dispute.opened":     true,
		"match.found":        true,
		"order.filled":       true,
		"transaction.created":   true,
		"transaction.delivered": true,
		"transaction.completed": true,
		"rating.submitted":      true,
	}

	for _, event := range req.Events {
		if !validEvents[event] {
			common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid event type: "+event))
			return
		}
	}

	webhook, err := h.repo.CreateWebhook(r.Context(), agent.ID, req.URL, req.Events)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to create webhook"))
		return
	}

	// Return with secret (only shown once)
	response := WebhookResponse{
		ID:        webhook.ID,
		URL:       webhook.URL,
		Secret:    webhook.Secret, // Only shown on create
		Events:    webhook.EventTypes,
		IsActive:  webhook.IsActive,
		CreatedAt: webhook.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	common.WriteJSON(w, http.StatusCreated, response)
}

// ListWebhooks handles GET /webhooks - list all webhooks for the authenticated agent.
func (h *WebhookHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	webhooks, err := h.repo.GetWebhooksByAgentID(r.Context(), agent.ID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to list webhooks"))
		return
	}

	responses := make([]WebhookResponse, len(webhooks))
	for i, wh := range webhooks {
		responses[i] = WebhookResponse{
			ID:        wh.ID,
			URL:       wh.URL,
			Events:    wh.EventTypes,
			IsActive:  wh.IsActive,
			CreatedAt: wh.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	common.WriteJSON(w, http.StatusOK, map[string]any{
		"webhooks": responses,
		"total":    len(responses),
	})
}

// DeleteWebhook handles DELETE /webhooks/{id} - delete a webhook.
func (h *WebhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid webhook id"))
		return
	}

	err = h.repo.DeleteWebhook(r.Context(), id, agent.ID)
	if err != nil {
		if err == notification.ErrWebhookNotFound {
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("webhook not found"))
			return
		}
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to delete webhook"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
