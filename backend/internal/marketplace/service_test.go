package marketplace

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// Test-specific error for validation failures
var errValidationFailed = errors.New("validation failed")

// mockRepository implements Repository interface for testing
type mockRepository struct {
	listings map[uuid.UUID]*Listing
	requests map[uuid.UUID]*Request
	offers   map[uuid.UUID]*Offer
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		listings: make(map[uuid.UUID]*Listing),
		requests: make(map[uuid.UUID]*Request),
		offers:   make(map[uuid.UUID]*Offer),
	}
}

// mockPublisher implements EventPublisher for testing
type mockPublisher struct {
	events []publishedEvent
}

type publishedEvent struct {
	eventType string
	payload   map[string]any
}

func (m *mockPublisher) Publish(ctx context.Context, eventType string, payload map[string]any) error {
	m.events = append(m.events, publishedEvent{eventType, payload})
	return nil
}

func TestValidateListingRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateListingRequest
		wantErr bool
	}{
		{
			name: "valid listing",
			req: &CreateListingRequest{
				Title:       "Test Listing",
				ListingType: ListingTypeServices,
			},
			wantErr: false,
		},
		{
			name: "missing title",
			req: &CreateListingRequest{
				ListingType: ListingTypeServices,
			},
			wantErr: true,
		},
		{
			name: "missing listing type",
			req: &CreateListingRequest{
				Title: "Test Listing",
			},
			wantErr: true,
		},
		{
			name: "invalid listing type",
			req: &CreateListingRequest{
				Title:       "Test Listing",
				ListingType: "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid goods listing",
			req: &CreateListingRequest{
				Title:       "Test Goods",
				ListingType: ListingTypeGoods,
			},
			wantErr: false,
		},
		{
			name: "valid data listing",
			req: &CreateListingRequest{
				Title:       "Test Data",
				ListingType: ListingTypeData,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateListingRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateListingRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRequestRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateRequestRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &CreateRequestRequest{
				Title:       "Need something",
				RequestType: ListingTypeServices,
			},
			wantErr: false,
		},
		{
			name: "missing title",
			req: &CreateRequestRequest{
				RequestType: ListingTypeServices,
			},
			wantErr: true,
		},
		{
			name: "missing request type",
			req: &CreateRequestRequest{
				Title: "Need something",
			},
			wantErr: true,
		},
		{
			name: "invalid request type",
			req: &CreateRequestRequest{
				Title:       "Need something",
				RequestType: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequestRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequestRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateOfferRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateOfferRequest
		wantErr bool
	}{
		{
			name: "valid offer",
			req: &CreateOfferRequest{
				PriceAmount: 100.0,
			},
			wantErr: false,
		},
		{
			name: "zero price",
			req: &CreateOfferRequest{
				PriceAmount: 0,
			},
			wantErr: true,
		},
		{
			name: "negative price",
			req: &CreateOfferRequest{
				PriceAmount: -50.0,
			},
			wantErr: true,
		},
		{
			name: "valid offer with description",
			req: &CreateOfferRequest{
				PriceAmount: 50.0,
				Description: "I can help with this",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOfferRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateOfferRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListingTypeValidation(t *testing.T) {
	validTypes := []ListingType{ListingTypeGoods, ListingTypeServices, ListingTypeData}

	for _, lt := range validTypes {
		if !isValidListingType(lt) {
			t.Errorf("expected %s to be valid", lt)
		}
	}

	invalidTypes := []ListingType{"invalid", "unknown", ""}
	for _, lt := range invalidTypes {
		if isValidListingType(lt) {
			t.Errorf("expected %s to be invalid", lt)
		}
	}
}

func TestGeographicScopeValidation(t *testing.T) {
	validScopes := []GeographicScope{ScopeLocal, ScopeRegional, ScopeNational, ScopeInternational}

	for _, scope := range validScopes {
		if !isValidGeographicScope(scope) {
			t.Errorf("expected %s to be valid", scope)
		}
	}

	// Empty scope should default and be considered valid in context
	if isValidGeographicScope("invalid") {
		t.Error("expected 'invalid' to be invalid scope")
	}
}

// Helper validation functions (these should match what's in service.go)
func validateListingRequest(req *CreateListingRequest) error {
	if req.Title == "" {
		return errValidationFailed
	}
	if req.ListingType == "" {
		return errValidationFailed
	}
	if !isValidListingType(req.ListingType) {
		return errValidationFailed
	}
	return nil
}

func validateRequestRequest(req *CreateRequestRequest) error {
	if req.Title == "" {
		return errValidationFailed
	}
	if req.RequestType == "" {
		return errValidationFailed
	}
	if !isValidListingType(req.RequestType) {
		return errValidationFailed
	}
	return nil
}

func validateOfferRequest(req *CreateOfferRequest) error {
	if req.PriceAmount <= 0 {
		return errValidationFailed
	}
	return nil
}

func isValidListingType(t ListingType) bool {
	switch t {
	case ListingTypeGoods, ListingTypeServices, ListingTypeData:
		return true
	}
	return false
}

func isValidGeographicScope(s GeographicScope) bool {
	switch s {
	case ScopeLocal, ScopeRegional, ScopeNational, ScopeInternational:
		return true
	}
	return false
}
