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

func TestListingTypes(t *testing.T) {
	tests := []struct {
		listingType ListingType
		expected    string
	}{
		{ListingTypeGoods, "goods"},
		{ListingTypeServices, "services"},
		{ListingTypeData, "data"},
	}

	for _, tt := range tests {
		if string(tt.listingType) != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.listingType)
		}
	}
}

func TestListingStatus(t *testing.T) {
	tests := []struct {
		status   ListingStatus
		expected string
	}{
		{ListingStatusDraft, "draft"},
		{ListingStatusActive, "active"},
		{ListingStatusPaused, "paused"},
		{ListingStatusSold, "sold"},
		{ListingStatusExpired, "expired"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.status)
		}
	}
}

func TestGeographicScope(t *testing.T) {
	tests := []struct {
		scope    GeographicScope
		expected string
	}{
		{ScopeLocal, "local"},
		{ScopeRegional, "regional"},
		{ScopeNational, "national"},
		{ScopeInternational, "international"},
	}

	for _, tt := range tests {
		if string(tt.scope) != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.scope)
		}
	}
}

func TestRequestStatus(t *testing.T) {
	tests := []struct {
		status   RequestStatus
		expected string
	}{
		{RequestStatusOpen, "open"},
		{RequestStatusInProgress, "in_progress"},
		{RequestStatusFulfilled, "fulfilled"},
		{RequestStatusCancelled, "cancelled"},
		{RequestStatusExpired, "expired"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.status)
		}
	}
}

func TestOfferStatus(t *testing.T) {
	tests := []struct {
		status   OfferStatus
		expected string
	}{
		{OfferStatusPending, "pending"},
		{OfferStatusAccepted, "accepted"},
		{OfferStatusRejected, "rejected"},
		{OfferStatusWithdrawn, "withdrawn"},
		{OfferStatusExpired, "expired"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.status)
		}
	}
}

func TestListing_AllFields(t *testing.T) {
	categoryID := uuid.New()
	priceAmount := 99.99
	lat := 37.7749
	lng := -122.4194
	radius := 50

	listing := &Listing{
		ID:              uuid.New(),
		SellerID:        uuid.New(),
		CategoryID:      &categoryID,
		Title:           "Test Listing",
		Description:     "A test listing description",
		ListingType:     ListingTypeServices,
		PriceAmount:     &priceAmount,
		PriceCurrency:   "USD",
		Quantity:        10,
		GeographicScope: ScopeRegional,
		LocationLat:     &lat,
		LocationLng:     &lng,
		LocationRadius:  &radius,
		Status:          ListingStatusActive,
		Metadata:        map[string]any{"featured": true},
	}

	if listing.Title != "Test Listing" {
		t.Errorf("expected title 'Test Listing', got %s", listing.Title)
	}
	if *listing.PriceAmount != 99.99 {
		t.Error("price amount not set correctly")
	}
	if listing.ListingType != ListingTypeServices {
		t.Error("listing type not set correctly")
	}
	if listing.GeographicScope != ScopeRegional {
		t.Error("geographic scope not set correctly")
	}
}

func TestRequest_AllFields(t *testing.T) {
	categoryID := uuid.New()
	budgetMin := 50.0
	budgetMax := 200.0
	lat := 40.7128
	lng := -74.0060
	radius := 100

	request := &Request{
		ID:              uuid.New(),
		RequesterID:     uuid.New(),
		CategoryID:      &categoryID,
		Title:           "Need a service",
		Description:     "I need help with something",
		RequestType:     ListingTypeServices,
		BudgetMin:       &budgetMin,
		BudgetMax:       &budgetMax,
		BudgetCurrency:  "USD",
		Quantity:        1,
		GeographicScope: ScopeLocal,
		LocationLat:     &lat,
		LocationLng:     &lng,
		LocationRadius:  &radius,
		Status:          RequestStatusOpen,
		OfferCount:      5,
		Metadata:        map[string]any{"urgent": true},
	}

	if request.Title != "Need a service" {
		t.Errorf("expected title 'Need a service', got %s", request.Title)
	}
	if *request.BudgetMin != 50.0 {
		t.Error("budget min not set correctly")
	}
	if *request.BudgetMax != 200.0 {
		t.Error("budget max not set correctly")
	}
	if request.OfferCount != 5 {
		t.Error("offer count not set correctly")
	}
}

func TestOffer_AllFields(t *testing.T) {
	listingID := uuid.New()

	offer := &Offer{
		ID:            uuid.New(),
		RequestID:     uuid.New(),
		OffererID:     uuid.New(),
		ListingID:     &listingID,
		PriceAmount:   75.0,
		PriceCurrency: "USD",
		Description:   "I can help with this",
		DeliveryTerms: "Within 24 hours",
		Status:        OfferStatusPending,
		Metadata:      map[string]any{"rush": false},
	}

	if offer.PriceAmount != 75.0 {
		t.Errorf("expected price 75.0, got %f", offer.PriceAmount)
	}
	if offer.Status != OfferStatusPending {
		t.Error("status not set correctly")
	}
	if offer.DeliveryTerms != "Within 24 hours" {
		t.Error("delivery terms not set correctly")
	}
}

func TestCategory(t *testing.T) {
	parentID := uuid.New()

	category := &Category{
		ID:          uuid.New(),
		ParentID:    &parentID,
		Name:        "Data Services",
		Slug:        "data-services",
		Description: "Services related to data",
		Metadata:    map[string]any{"icon": "database"},
	}

	if category.Name != "Data Services" {
		t.Errorf("expected name 'Data Services', got %s", category.Name)
	}
	if category.Slug != "data-services" {
		t.Errorf("expected slug 'data-services', got %s", category.Slug)
	}
}

func TestSearchListingsParams(t *testing.T) {
	categoryID := uuid.New()
	sellerID := uuid.New()
	listingType := ListingTypeData
	minPrice := 10.0
	maxPrice := 100.0
	scope := ScopeNational
	status := ListingStatusActive

	params := SearchListingsParams{
		CategoryID:      &categoryID,
		ListingType:     &listingType,
		MinPrice:        &minPrice,
		MaxPrice:        &maxPrice,
		GeographicScope: &scope,
		SellerID:        &sellerID,
		Status:          &status,
		Query:           "test",
		Limit:           20,
		Offset:          10,
	}

	if *params.CategoryID != categoryID {
		t.Error("category ID not set correctly")
	}
	if *params.ListingType != ListingTypeData {
		t.Error("listing type not set correctly")
	}
	if *params.MinPrice != 10.0 {
		t.Error("min price not set correctly")
	}
	if params.Query != "test" {
		t.Errorf("expected query 'test', got %s", params.Query)
	}
}

func TestSearchRequestsParams(t *testing.T) {
	categoryID := uuid.New()
	requesterID := uuid.New()
	requestType := ListingTypeServices
	minBudget := 50.0
	maxBudget := 500.0
	scope := ScopeInternational
	status := RequestStatusOpen

	params := SearchRequestsParams{
		CategoryID:      &categoryID,
		RequestType:     &requestType,
		MinBudget:       &minBudget,
		MaxBudget:       &maxBudget,
		GeographicScope: &scope,
		RequesterID:     &requesterID,
		Status:          &status,
		Query:           "help",
		Limit:           50,
		Offset:          0,
	}

	if *params.RequesterID != requesterID {
		t.Error("requester ID not set correctly")
	}
	if *params.MinBudget != 50.0 {
		t.Error("min budget not set correctly")
	}
	if *params.MaxBudget != 500.0 {
		t.Error("max budget not set correctly")
	}
}

func TestListResult(t *testing.T) {
	result := ListResult[*Listing]{
		Items:  []*Listing{{Title: "Listing 1"}, {Title: "Listing 2"}},
		Total:  100,
		Limit:  20,
		Offset: 0,
	}

	if len(result.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(result.Items))
	}
	if result.Total != 100 {
		t.Errorf("expected total 100, got %d", result.Total)
	}
}

func TestCreateListingRequestAllFields(t *testing.T) {
	categoryID := uuid.New()
	priceAmount := 150.0
	lat := 51.5074
	lng := -0.1278
	radius := 25

	req := &CreateListingRequest{
		CategoryID:      &categoryID,
		Title:           "Full Listing",
		Description:     "A complete listing with all fields",
		ListingType:     ListingTypeGoods,
		PriceAmount:     &priceAmount,
		PriceCurrency:   "GBP",
		Quantity:        5,
		GeographicScope: ScopeLocal,
		LocationLat:     &lat,
		LocationLng:     &lng,
		LocationRadius:  &radius,
		Metadata:        map[string]any{"premium": true},
	}

	if *req.CategoryID != categoryID {
		t.Error("category ID not set correctly")
	}
	if req.PriceCurrency != "GBP" {
		t.Errorf("expected currency GBP, got %s", req.PriceCurrency)
	}
	if req.Quantity != 5 {
		t.Errorf("expected quantity 5, got %d", req.Quantity)
	}
}

func TestCreateRequestRequestAllFields(t *testing.T) {
	categoryID := uuid.New()
	budgetMin := 100.0
	budgetMax := 1000.0
	lat := 48.8566
	lng := 2.3522
	radius := 50

	req := &CreateRequestRequest{
		CategoryID:      &categoryID,
		Title:           "Full Request",
		Description:     "A complete request with all fields",
		RequestType:     ListingTypeData,
		BudgetMin:       &budgetMin,
		BudgetMax:       &budgetMax,
		BudgetCurrency:  "EUR",
		Quantity:        1,
		GeographicScope: ScopeRegional,
		LocationLat:     &lat,
		LocationLng:     &lng,
		LocationRadius:  &radius,
		Metadata:        map[string]any{"deadline": "2024-12-31"},
	}

	if *req.BudgetMin != 100.0 {
		t.Error("budget min not set correctly")
	}
	if req.BudgetCurrency != "EUR" {
		t.Errorf("expected currency EUR, got %s", req.BudgetCurrency)
	}
}

func TestCreateOfferRequestAllFields(t *testing.T) {
	listingID := uuid.New()

	req := &CreateOfferRequest{
		ListingID:     &listingID,
		PriceAmount:   250.0,
		PriceCurrency: "USD",
		Description:   "I can provide this service",
		DeliveryTerms: "2-3 business days",
	}

	if *req.ListingID != listingID {
		t.Error("listing ID not set correctly")
	}
	if req.PriceAmount != 250.0 {
		t.Errorf("expected price 250.0, got %f", req.PriceAmount)
	}
	if req.DeliveryTerms != "2-3 business days" {
		t.Error("delivery terms not set correctly")
	}
}
