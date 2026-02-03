package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/digi604/swarmmarket/backend/internal/common"
	"github.com/digi604/swarmmarket/backend/internal/task"
	"github.com/digi604/swarmmarket/backend/pkg/middleware"
)

// TaskHandler handles task HTTP requests.
type TaskHandler struct {
	service *task.Service
}

// NewTaskHandler creates a new task handler.
func NewTaskHandler(service *task.Service) *TaskHandler {
	return &TaskHandler{service: service}
}

// CreateTask handles POST /api/v1/tasks
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	var req task.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	t, err := h.service.CreateTask(r.Context(), agent.ID, &req)
	if err != nil {
		handleTaskError(w, err)
		return
	}

	common.WriteJSON(w, http.StatusCreated, t)
}

// GetTask handles GET /api/v1/tasks/{taskId}
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "taskId"))
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid task id"))
		return
	}

	t, err := h.service.GetTask(r.Context(), taskID)
	if err != nil {
		handleTaskError(w, err)
		return
	}

	// Check authorization - only requester or executor can view
	agent := middleware.GetAgent(r.Context())
	if agent == nil || (agent.ID != t.RequesterID && agent.ID != t.ExecutorID) {
		common.WriteError(w, http.StatusForbidden, common.ErrForbidden("not authorized to view this task"))
		return
	}

	common.WriteJSON(w, http.StatusOK, t)
}

// ListTasks handles GET /api/v1/tasks
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	params := task.ListTasksParams{
		Limit:  parseIntQueryParam(r, "limit", 20),
		Offset: parseIntQueryParam(r, "offset", 0),
	}

	// Filter by role
	role := r.URL.Query().Get("role")
	switch role {
	case "requester":
		params.RequesterID = &agent.ID
	case "executor":
		params.ExecutorID = &agent.ID
	default:
		// Show tasks where agent is either requester or executor
		params.AgentID = &agent.ID
	}

	// Filter by status
	if status := r.URL.Query().Get("status"); status != "" {
		s := task.TaskStatus(status)
		params.Status = &s
	}

	// Filter by capability
	if capID := r.URL.Query().Get("capability_id"); capID != "" {
		if id, err := uuid.Parse(capID); err == nil {
			params.CapabilityID = &id
		}
	}

	result, err := h.service.ListTasks(r.Context(), params)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to list tasks"))
		return
	}

	common.WriteJSON(w, http.StatusOK, result)
}

// GetTaskHistory handles GET /api/v1/tasks/{taskId}/history
func (h *TaskHandler) GetTaskHistory(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "taskId"))
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid task id"))
		return
	}

	// First check if user can access this task
	t, err := h.service.GetTask(r.Context(), taskID)
	if err != nil {
		handleTaskError(w, err)
		return
	}

	agent := middleware.GetAgent(r.Context())
	if agent == nil || (agent.ID != t.RequesterID && agent.ID != t.ExecutorID) {
		common.WriteError(w, http.StatusForbidden, common.ErrForbidden("not authorized to view this task"))
		return
	}

	history, err := h.service.GetTaskHistory(r.Context(), taskID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to get history"))
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]any{"history": history})
}

// AcceptTask handles POST /api/v1/tasks/{taskId}/accept
func (h *TaskHandler) AcceptTask(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	taskID, err := uuid.Parse(chi.URLParam(r, "taskId"))
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid task id"))
		return
	}

	t, err := h.service.AcceptTask(r.Context(), agent.ID, taskID)
	if err != nil {
		handleTaskError(w, err)
		return
	}

	common.WriteJSON(w, http.StatusOK, t)
}

// UpdateProgress handles POST /api/v1/tasks/{taskId}/progress
func (h *TaskHandler) UpdateProgress(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	taskID, err := uuid.Parse(chi.URLParam(r, "taskId"))
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid task id"))
		return
	}

	var req task.UpdateTaskProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	t, err := h.service.UpdateTaskProgress(r.Context(), agent.ID, taskID, &req)
	if err != nil {
		handleTaskError(w, err)
		return
	}

	common.WriteJSON(w, http.StatusOK, t)
}

// DeliverTask handles POST /api/v1/tasks/{taskId}/deliver
func (h *TaskHandler) DeliverTask(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	taskID, err := uuid.Parse(chi.URLParam(r, "taskId"))
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid task id"))
		return
	}

	var req task.DeliverTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	t, err := h.service.DeliverTask(r.Context(), agent.ID, taskID, &req)
	if err != nil {
		handleTaskError(w, err)
		return
	}

	common.WriteJSON(w, http.StatusOK, t)
}

// ConfirmTask handles POST /api/v1/tasks/{taskId}/confirm
func (h *TaskHandler) ConfirmTask(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	taskID, err := uuid.Parse(chi.URLParam(r, "taskId"))
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid task id"))
		return
	}

	t, err := h.service.ConfirmTask(r.Context(), agent.ID, taskID)
	if err != nil {
		handleTaskError(w, err)
		return
	}

	common.WriteJSON(w, http.StatusOK, t)
}

// CancelTask handles POST /api/v1/tasks/{taskId}/cancel
func (h *TaskHandler) CancelTask(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	taskID, err := uuid.Parse(chi.URLParam(r, "taskId"))
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid task id"))
		return
	}

	t, err := h.service.CancelTask(r.Context(), agent.ID, taskID)
	if err != nil {
		handleTaskError(w, err)
		return
	}

	common.WriteJSON(w, http.StatusOK, t)
}

// FailTask handles POST /api/v1/tasks/{taskId}/fail
func (h *TaskHandler) FailTask(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	taskID, err := uuid.Parse(chi.URLParam(r, "taskId"))
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid task id"))
		return
	}

	var req task.FailTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	t, err := h.service.FailTask(r.Context(), agent.ID, taskID, &req)
	if err != nil {
		handleTaskError(w, err)
		return
	}

	common.WriteJSON(w, http.StatusOK, t)
}

// handleTaskError converts task errors to HTTP responses.
func handleTaskError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, task.ErrTaskNotFound):
		common.WriteError(w, http.StatusNotFound, common.ErrNotFound("task not found"))
	case errors.Is(err, task.ErrNotAuthorized):
		common.WriteError(w, http.StatusForbidden, common.ErrForbidden("not authorized"))
	case errors.Is(err, task.ErrInvalidStatus):
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest(err.Error()))
	case errors.Is(err, task.ErrCapabilityNotFound):
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("capability not found"))
	case errors.Is(err, task.ErrCapabilityInactive):
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("capability is not accepting tasks"))
	case errors.Is(err, task.ErrInputValidation):
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest(err.Error()))
	case errors.Is(err, task.ErrOutputValidation):
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest(err.Error()))
	case errors.Is(err, task.ErrInvalidEvent):
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest(err.Error()))
	case errors.Is(err, task.ErrSelfAssignment):
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("cannot create task for your own capability"))
	default:
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer(err.Error()))
	}
}

// parseIntQueryParam parses an integer query parameter with a default value.
func parseIntQueryParam(r *http.Request, name string, defaultVal int) int {
	val := r.URL.Query().Get(name)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}
