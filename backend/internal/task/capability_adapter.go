package task

import (
	"context"

	"github.com/google/uuid"

	"github.com/digi604/swarmmarket/backend/internal/capability"
)

// CapabilityAdapter adapts the capability.Service to the CapabilityGetter interface.
type CapabilityAdapter struct {
	service *capability.Service
}

// NewCapabilityAdapter creates a new capability adapter.
func NewCapabilityAdapter(service *capability.Service) *CapabilityAdapter {
	return &CapabilityAdapter{service: service}
}

// GetCapabilityByID retrieves capability information needed for task validation.
func (a *CapabilityAdapter) GetCapabilityByID(ctx context.Context, id uuid.UUID) (*CapabilityInfo, error) {
	cap, err := a.service.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	info := &CapabilityInfo{
		ID:               cap.ID,
		AgentID:          cap.AgentID,
		Name:             cap.Name,
		InputSchema:      cap.InputSchema,
		OutputSchema:     cap.OutputSchema,
		StatusEvents:     cap.StatusEvents,
		IsActive:         cap.IsActive,
		IsAcceptingTasks: cap.IsAcceptingTasks,
		BaseFee:          cap.BaseFee,
		PercentageFee:    cap.PercentageFee,
		Currency:         cap.Currency,
		PricingModel:     string(cap.PricingModel),
	}

	return info, nil
}

// CapabilityStatsAdapter adapts capability.Service to CapabilityStatsUpdater.
type CapabilityStatsAdapter struct {
	service *capability.Service
}

// NewCapabilityStatsAdapter creates a new capability stats adapter.
func NewCapabilityStatsAdapter(service *capability.Service) *CapabilityStatsAdapter {
	return &CapabilityStatsAdapter{service: service}
}

// RecordTaskCompletion records task completion statistics.
func (a *CapabilityStatsAdapter) RecordTaskCompletion(ctx context.Context, capabilityID uuid.UUID, success bool, rating *float64) error {
	return a.service.RecordTaskCompletion(ctx, capabilityID, success, rating)
}
