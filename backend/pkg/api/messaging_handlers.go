package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/digi604/swarmmarket/backend/internal/common"
	"github.com/digi604/swarmmarket/backend/internal/messaging"
	"github.com/digi604/swarmmarket/backend/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// MessagingHandler handles messaging-related HTTP requests.
type MessagingHandler struct {
	service *messaging.Service
}

// NewMessagingHandler creates a new messaging handler.
func NewMessagingHandler(service *messaging.Service) *MessagingHandler {
	return &MessagingHandler{service: service}
}

// SendMessage handles POST /api/v1/messages
func (h *MessagingHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	var req messaging.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	msg, conv, err := h.service.SendMessage(r.Context(), agent.ID, &req)
	if err != nil {
		log.Printf("[ERROR] SendMessage failed: %v (agentID=%s)", err, agent.ID)
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest(err.Error()))
		return
	}

	common.WriteJSON(w, http.StatusCreated, map[string]any{
		"message":      msg,
		"conversation": conv,
	})
}

// ListConversations handles GET /api/v1/conversations
func (h *MessagingHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	limit := parseIntParam(r, "limit", 20)
	offset := parseIntParam(r, "offset", 0)

	result, err := h.service.GetConversations(r.Context(), agent.ID, messaging.ListConversationsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		log.Printf("[ERROR] ListConversations failed: %v (agentID=%s)", err, agent.ID)
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to get conversations"))
		return
	}

	common.WriteJSON(w, http.StatusOK, result)
}

// GetConversation handles GET /api/v1/conversations/{id}
func (h *MessagingHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	conversationIDStr := chi.URLParam(r, "id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid conversation id"))
		return
	}

	conv, err := h.service.GetConversation(r.Context(), agent.ID, conversationID)
	if err != nil {
		log.Printf("[ERROR] GetConversation failed: %v (conversationID=%s)", err, conversationID)
		common.WriteError(w, http.StatusNotFound, common.ErrNotFound("conversation not found"))
		return
	}

	common.WriteJSON(w, http.StatusOK, conv)
}

// GetMessages handles GET /api/v1/conversations/{id}/messages
func (h *MessagingHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	conversationIDStr := chi.URLParam(r, "id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid conversation id"))
		return
	}

	limit := parseIntParam(r, "limit", 50)
	offset := parseIntParam(r, "offset", 0)

	result, err := h.service.GetMessages(r.Context(), agent.ID, conversationID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetMessages failed: %v (conversationID=%s)", err, conversationID)
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest(err.Error()))
		return
	}

	common.WriteJSON(w, http.StatusOK, result)
}

// ReplyToConversation handles POST /api/v1/conversations/{id}/messages
func (h *MessagingHandler) ReplyToConversation(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	conversationIDStr := chi.URLParam(r, "id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid conversation id"))
		return
	}

	var req messaging.ReplyToConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	msg, err := h.service.ReplyToConversation(r.Context(), agent.ID, conversationID, &req)
	if err != nil {
		log.Printf("[ERROR] ReplyToConversation failed: %v (conversationID=%s)", err, conversationID)
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest(err.Error()))
		return
	}

	common.WriteJSON(w, http.StatusCreated, msg)
}

// MarkAsRead handles POST /api/v1/conversations/{id}/read
func (h *MessagingHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	conversationIDStr := chi.URLParam(r, "id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid conversation id"))
		return
	}

	if err := h.service.MarkAsRead(r.Context(), agent.ID, conversationID); err != nil {
		log.Printf("[ERROR] MarkAsRead failed: %v (conversationID=%s)", err, conversationID)
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest(err.Error()))
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]string{"status": "read"})
}

// GetUnreadCount handles GET /api/v1/messages/unread-count
func (h *MessagingHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	count, err := h.service.GetUnreadCount(r.Context(), agent.ID)
	if err != nil {
		log.Printf("[ERROR] GetUnreadCount failed: %v (agentID=%s)", err, agent.ID)
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to get unread count"))
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]int{"unread_count": count})
}

// Note: parseIntParam is defined in marketplace_handlers.go
