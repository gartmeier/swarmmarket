package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/digi604/swarmmarket/backend/internal/agent"
	"github.com/digi604/swarmmarket/backend/internal/common"
	"github.com/digi604/swarmmarket/backend/internal/trust"
	"github.com/digi604/swarmmarket/backend/pkg/middleware"
)

// TrustHandler handles trust-related API endpoints.
type TrustHandler struct {
	trustService *trust.Service
}

// NewTrustHandler creates a new trust handler.
func NewTrustHandler(trustService *trust.Service) *TrustHandler {
	return &TrustHandler{
		trustService: trustService,
	}
}

// GetTrustBreakdown returns the authenticated agent's trust score breakdown.
func (h *TrustHandler) GetTrustBreakdown(w http.ResponseWriter, r *http.Request) {
	ag := r.Context().Value(middleware.AgentContextKey).(*agent.Agent)

	breakdown, err := h.trustService.GetTrustBreakdown(r.Context(), ag.ID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer(err.Error()))
		return
	}

	common.WriteJSON(w, http.StatusOK, breakdown)
}

// GetAgentTrustBreakdown returns any agent's trust score breakdown.
func (h *TrustHandler) GetAgentTrustBreakdown(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	agentID, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid agent id"))
		return
	}

	breakdown, err := h.trustService.GetTrustBreakdown(r.Context(), agentID)
	if err != nil {
		common.WriteError(w, http.StatusNotFound, common.ErrNotFound("agent not found"))
		return
	}

	common.WriteJSON(w, http.StatusOK, breakdown)
}

// GetTrustHistory returns the trust score change history for an agent.
func (h *TrustHandler) GetTrustHistory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	agentID, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid agent id"))
		return
	}

	// Default limit
	limit := 50

	history, err := h.trustService.GetTrustHistory(r.Context(), agentID, limit)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer(err.Error()))
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]any{
		"agent_id": agentID,
		"history":  history,
	})
}

// InitiateTwitterVerification starts the Twitter verification process.
func (h *TrustHandler) InitiateTwitterVerification(w http.ResponseWriter, r *http.Request) {
	ag := r.Context().Value(middleware.AgentContextKey).(*agent.Agent)

	resp, err := h.trustService.InitiateTwitterVerification(r.Context(), ag.ID, ag.Name)
	if err != nil {
		if err == trust.ErrAlreadyVerified {
			common.WriteError(w, http.StatusConflict, common.ErrConflict("Twitter is already verified for this agent"))
			return
		}
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer(err.Error()))
		return
	}

	common.WriteJSON(w, http.StatusOK, resp)
}

// ConfirmTwitterVerification confirms Twitter verification with a tweet URL.
func (h *TrustHandler) ConfirmTwitterVerification(w http.ResponseWriter, r *http.Request) {
	ag := r.Context().Value(middleware.AgentContextKey).(*agent.Agent)

	var req trust.ConfirmTwitterVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	if req.ChallengeID == "" {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("challenge_id is required"))
		return
	}
	if req.TweetURL == "" {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("tweet_url is required"))
		return
	}

	resp, err := h.trustService.ConfirmTwitterVerification(r.Context(), ag.ID, &req)
	if err != nil {
		switch err {
		case trust.ErrTwitterNotConfigured:
			common.WriteError(w, http.StatusServiceUnavailable, common.ErrServiceUnavailable("Twitter verification is not configured"))
		case trust.ErrInvalidChallengeID:
			common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid challenge id"))
		case trust.ErrChallengeNotFound:
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("challenge not found"))
		case trust.ErrChallengeExpired:
			common.WriteError(w, http.StatusGone, common.ErrGone("challenge has expired"))
		case trust.ErrChallengeNotPending:
			common.WriteError(w, http.StatusConflict, common.ErrConflict("challenge is not pending"))
		case trust.ErrMaxAttemptsExceeded:
			common.WriteError(w, http.StatusTooManyRequests, common.ErrTooManyRequests("maximum verification attempts exceeded"))
		case trust.ErrUnauthorized:
			common.WriteError(w, http.StatusForbidden, common.ErrForbidden("you don't own this challenge"))
		default:
			common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer(err.Error()))
		}
		return
	}

	common.WriteJSON(w, http.StatusOK, resp)
}

// ListVerifications returns all verifications for the authenticated agent.
func (h *TrustHandler) ListVerifications(w http.ResponseWriter, r *http.Request) {
	ag := r.Context().Value(middleware.AgentContextKey).(*agent.Agent)

	verifications, err := h.trustService.GetVerifications(r.Context(), ag.ID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer(err.Error()))
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]any{
		"verifications": verifications,
	})
}
