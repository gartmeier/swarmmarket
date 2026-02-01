package api

import (
	"encoding/json"
	"net/http"

	"github.com/digi604/swarmmarket/backend/internal/agent"
	"github.com/digi604/swarmmarket/backend/internal/common"
	"github.com/digi604/swarmmarket/backend/internal/user"
	"github.com/digi604/swarmmarket/backend/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// DashboardHandler handles dashboard API requests for human users.
type DashboardHandler struct {
	userService  *user.Service
	agentService *agent.Service
}

// NewDashboardHandler creates a new dashboard handler.
func NewDashboardHandler(userService *user.Service, agentService *agent.Service) *DashboardHandler {
	return &DashboardHandler{
		userService:  userService,
		agentService: agentService,
	}
}

// ListOwnedAgents returns all agents owned by the authenticated user.
// GET /api/v1/dashboard/agents
func (h *DashboardHandler) ListOwnedAgents(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUser(r.Context())
	if usr == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	agents, err := h.userService.GetOwnedAgents(r.Context(), usr.ID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to get owned agents"))
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]any{
		"agents": agents,
		"count":  len(agents),
	})
}

// GetAgentMetrics returns detailed metrics for an owned agent.
// GET /api/v1/dashboard/agents/{id}/metrics
func (h *DashboardHandler) GetAgentMetrics(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUser(r.Context())
	if usr == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	agentIDStr := chi.URLParam(r, "id")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid agent id"))
		return
	}

	metrics, err := h.userService.GetAgentMetrics(r.Context(), usr.ID, agentID)
	if err != nil {
		if err.Error() == "agent not found" {
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("agent not found"))
			return
		}
		if err.Error() == "not authorized to view this agent's metrics" {
			common.WriteError(w, http.StatusForbidden, common.ErrForbidden("not authorized to view this agent's metrics"))
			return
		}
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to get agent metrics"))
		return
	}

	common.WriteJSON(w, http.StatusOK, metrics)
}

// ClaimAgentOwnership claims ownership of an agent using an ownership token.
// POST /api/v1/dashboard/agents/claim
func (h *DashboardHandler) ClaimAgentOwnership(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUser(r.Context())
	if usr == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	var req user.ClaimOwnershipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	if req.Token == "" {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("token is required"))
		return
	}

	claimedAgent, err := h.agentService.ClaimOwnership(r.Context(), usr.ID, req.Token)
	if err != nil {
		switch err {
		case agent.ErrTokenNotFound:
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("invalid ownership token"))
		case agent.ErrTokenExpired:
			common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("ownership token has expired"))
		case agent.ErrTokenAlreadyUsed:
			common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("ownership token has already been used"))
		case agent.ErrAgentAlreadyOwned:
			common.WriteError(w, http.StatusConflict, common.ErrConflict("agent already has an owner"))
		default:
			common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to claim ownership"))
		}
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "ownership claimed successfully",
		"agent":   claimedAgent.PublicProfile(),
	})
}

// GetProfile returns the authenticated user's profile.
// GET /api/v1/dashboard/profile
func (h *DashboardHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	usr := middleware.GetUser(r.Context())
	if usr == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	common.WriteJSON(w, http.StatusOK, usr)
}
