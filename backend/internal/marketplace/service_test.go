package marketplace

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// Test-specific error for validation failures
var errValidationFailed = errors.New("validation failed")

// mockRepository implements RepositoryInterface for testing
type mockRepository struct {
	listings         map[uuid.UUID]*Listing
	requests         map[uuid.UUID]*Request
	offers           map[uuid.UUID]*Offer
	categories       []Category
	createListingErr error
	createRequestErr error
	createOfferErr   error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		listings: make(map[uuid.UUID]*Listing),
		requests: make(map[uuid.UUID]*Request),
		offers:   make(map[uuid.UUID]*Offer),
	}
}

// Verify mockRepository implements RepositoryInterface
var _ RepositoryInterface = (*mockRepository)(nil)

func (m *mockRepository) CreateListing(ctx context.Context, listing *Listing) error {
	if m.createListingErr != nil {
		return m.createListingErr
	}
	m.listings[listing.ID] = listing
	return nil
}

func (m *mockRepository) GetListingByID(ctx context.Context, id uuid.UUID) (*Listing, error) {
	listing, ok := m.listings[id]
	if !ok {
		return nil, ErrListingNotFound
	}
	return listing, nil
}

func (m *mockRepository) SearchListings(ctx context.Context, params SearchListingsParams) (*ListResult[Listing], error) {
	var items []Listing
	for _, l := range m.listings {
		if params.Status != nil && l.Status != *params.Status {
			continue
		}
		if params.SellerID != nil && l.SellerID != *params.SellerID {
			continue
		}
		items = append(items, *l)
	}
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	return &ListResult[Listing]{
		Items:  items,
		Total:  len(items),
		Limit:  limit,
		Offset: params.Offset,
	}, nil
}

func (m *mockRepository) DeleteListing(ctx context.Context, id uuid.UUID, sellerID uuid.UUID) error {
	listing, ok := m.listings[id]
	if !ok || listing.SellerID != sellerID {
		return ErrListingNotFound
	}
	listing.Status = ListingStatusExpired
	return nil
}

func (m *mockRepository) CreateRequest(ctx context.Context, req *Request) error {
	if m.createRequestErr != nil {
		return m.createRequestErr
	}
	m.requests[req.ID] = req
	return nil
}

func (m *mockRepository) GetRequestByID(ctx context.Context, id uuid.UUID) (*Request, error) {
	req, ok := m.requests[id]
	if !ok {
		return nil, ErrRequestNotFound
	}
	// Count offers for this request
	for _, o := range m.offers {
		if o.RequestID == id {
			req.OfferCount++
		}
	}
	return req, nil
}

func (m *mockRepository) SearchRequests(ctx context.Context, params SearchRequestsParams) (*ListResult[Request], error) {
	var items []Request
	for _, r := range m.requests {
		if params.Status != nil && r.Status != *params.Status {
			continue
		}
		if params.RequesterID != nil && r.RequesterID != *params.RequesterID {
			continue
		}
		items = append(items, *r)
	}
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	return &ListResult[Request]{
		Items:  items,
		Total:  len(items),
		Limit:  limit,
		Offset: params.Offset,
	}, nil
}

func (m *mockRepository) CreateOffer(ctx context.Context, offer *Offer) error {
	if m.createOfferErr != nil {
		return m.createOfferErr
	}
	m.offers[offer.ID] = offer
	return nil
}

func (m *mockRepository) GetOfferByID(ctx context.Context, id uuid.UUID) (*Offer, error) {
	offer, ok := m.offers[id]
	if !ok {
		return nil, ErrOfferNotFound
	}
	return offer, nil
}

func (m *mockRepository) GetOffersByRequestID(ctx context.Context, requestID uuid.UUID) ([]Offer, error) {
	var offers []Offer
	for _, o := range m.offers {
		if o.RequestID == requestID {
			offers = append(offers, *o)
		}
	}
	return offers, nil
}

func (m *mockRepository) UpdateOfferStatus(ctx context.Context, id uuid.UUID, status OfferStatus) error {
	offer, ok := m.offers[id]
	if !ok {
		return ErrOfferNotFound
	}
	offer.Status = status
	return nil
}

func (m *mockRepository) GetCategories(ctx context.Context) ([]Category, error) {
	return m.categories, nil
}

func (m *mockRepository) GetListingBySlug(ctx context.Context, slug string) (*Listing, error) {
	for _, l := range m.listings {
		if l.Slug == slug {
			return l, nil
		}
	}
	return nil, ErrListingNotFound
}

func (m *mockRepository) GetRequestBySlug(ctx context.Context, slug string) (*Request, error) {
	for _, r := range m.requests {
		if r.Slug == slug {
			return r, nil
		}
	}
	return nil, ErrRequestNotFound
}

func (m *mockRepository) UpdateRequest(ctx context.Context, id uuid.UUID, requesterID uuid.UUID, updates *UpdateRequestRequest) (*Request, error) {
	req, ok := m.requests[id]
	if !ok {
		return nil, ErrRequestNotFound
	}
	if req.RequesterID != requesterID {
		return nil, ErrRequestNotFound
	}
	if req.Status != RequestStatusOpen {
		return nil, ErrRequestNotFound
	}
	if updates.Title != nil {
		req.Title = *updates.Title
	}
	if updates.Description != nil {
		req.Description = *updates.Description
	}
	return req, nil
}

func (m *mockRepository) CreateComment(ctx context.Context, comment *Comment) error {
	return nil
}

func (m *mockRepository) GetCommentsByListingID(ctx context.Context, listingID uuid.UUID, limit, offset int) ([]Comment, int, error) {
	return nil, 0, nil
}

func (m *mockRepository) GetCommentReplies(ctx context.Context, parentID uuid.UUID) ([]Comment, error) {
	return nil, nil
}

func (m *mockRepository) DeleteComment(ctx context.Context, commentID, agentID uuid.UUID) error {
	return nil
}

func (m *mockRepository) CreateRequestComment(ctx context.Context, comment *Comment) error {
	return nil
}

func (m *mockRepository) GetCommentsByRequestID(ctx context.Context, requestID uuid.UUID, limit, offset int) ([]Comment, int, error) {
	return nil, 0, nil
}

func (m *mockRepository) GetRequestCommentReplies(ctx context.Context, parentID uuid.UUID) ([]Comment, error) {
	return nil, nil
}

func (m *mockRepository) DeleteRequestComment(ctx context.Context, commentID, agentID uuid.UUID) error {
	return nil
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

// --- Mock Transaction Creator ---

type mockTransactionCreator struct {
	transactions map[uuid.UUID]struct {
		buyerID, sellerID uuid.UUID
		amount            float64
	}
	err error
}

func newMockTransactionCreator() *mockTransactionCreator {
	return &mockTransactionCreator{
		transactions: make(map[uuid.UUID]struct {
			buyerID, sellerID uuid.UUID
			amount            float64
		}),
	}
}

func (m *mockTransactionCreator) CreateFromOffer(ctx context.Context, buyerID, sellerID uuid.UUID, requestID, offerID *uuid.UUID, amount float64, currency string) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	txID := uuid.New()
	m.transactions[txID] = struct {
		buyerID, sellerID uuid.UUID
		amount            float64
	}{buyerID, sellerID, amount}
	return txID, nil
}

// --- Service Method Tests ---

func TestService_CreateListing(t *testing.T) {
	repo := newMockRepository()
	publisher := &mockPublisher{}
	service := NewService(repo, publisher)

	sellerID := uuid.New()
	priceAmount := 100.0

	listing, err := service.CreateListing(context.Background(), sellerID, &CreateListingRequest{
		Title:       "Test Listing",
		Description: "A test listing",
		ListingType: ListingTypeServices,
		PriceAmount: &priceAmount,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if listing.Title != "Test Listing" {
		t.Errorf("expected title 'Test Listing', got %s", listing.Title)
	}
	if listing.SellerID != sellerID {
		t.Errorf("expected seller ID %s, got %s", sellerID, listing.SellerID)
	}
	if listing.Status != ListingStatusActive {
		t.Errorf("expected status active, got %s", listing.Status)
	}
	if listing.PriceCurrency != "USD" {
		t.Errorf("expected currency USD, got %s", listing.PriceCurrency)
	}
	if listing.GeographicScope != ScopeInternational {
		t.Errorf("expected scope international, got %s", listing.GeographicScope)
	}
}

func TestService_CreateListing_MissingTitle(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	_, err := service.CreateListing(context.Background(), uuid.New(), &CreateListingRequest{
		ListingType: ListingTypeServices,
	})

	if err == nil {
		t.Error("expected error for missing title")
	}
}

func TestService_CreateListing_MissingType(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	_, err := service.CreateListing(context.Background(), uuid.New(), &CreateListingRequest{
		Title: "Test",
	})

	if err == nil {
		t.Error("expected error for missing listing type")
	}
}

func TestService_GetListing(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	// Create a listing first
	sellerID := uuid.New()
	listing, _ := service.CreateListing(context.Background(), sellerID, &CreateListingRequest{
		Title:       "Test Listing",
		ListingType: ListingTypeGoods,
	})

	// Get it back
	retrieved, err := service.GetListing(context.Background(), listing.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.ID != listing.ID {
		t.Errorf("expected ID %s, got %s", listing.ID, retrieved.ID)
	}
}

func TestService_GetListing_NotFound(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	_, err := service.GetListing(context.Background(), uuid.New())
	if err != ErrListingNotFound {
		t.Errorf("expected ErrListingNotFound, got %v", err)
	}
}

func TestService_SearchListings(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	sellerID := uuid.New()

	// Create some listings
	for i := 0; i < 5; i++ {
		service.CreateListing(context.Background(), sellerID, &CreateListingRequest{
			Title:       "Test Listing",
			ListingType: ListingTypeServices,
		})
	}

	// Search all
	result, err := service.SearchListings(context.Background(), SearchListingsParams{
		SellerID: &sellerID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 5 {
		t.Errorf("expected 5 listings, got %d", result.Total)
	}
}

func TestService_DeleteListing(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	sellerID := uuid.New()
	listing, _ := service.CreateListing(context.Background(), sellerID, &CreateListingRequest{
		Title:       "To Delete",
		ListingType: ListingTypeData,
	})

	err := service.DeleteListing(context.Background(), listing.ID, sellerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify status changed
	updated, _ := repo.GetListingByID(context.Background(), listing.ID)
	if updated.Status != ListingStatusExpired {
		t.Errorf("expected status expired, got %s", updated.Status)
	}
}

func TestService_DeleteListing_WrongSeller(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	sellerID := uuid.New()
	listing, _ := service.CreateListing(context.Background(), sellerID, &CreateListingRequest{
		Title:       "To Delete",
		ListingType: ListingTypeData,
	})

	err := service.DeleteListing(context.Background(), listing.ID, uuid.New()) // Different seller
	if err != ErrListingNotFound {
		t.Errorf("expected ErrListingNotFound, got %v", err)
	}
}

func TestService_CreateRequest(t *testing.T) {
	repo := newMockRepository()
	publisher := &mockPublisher{}
	service := NewService(repo, publisher)

	requesterID := uuid.New()
	budgetMin := 50.0
	budgetMax := 200.0

	request, err := service.CreateRequest(context.Background(), requesterID, &CreateRequestRequest{
		Title:       "Need Help",
		Description: "I need help with something",
		RequestType: ListingTypeServices,
		BudgetMin:   &budgetMin,
		BudgetMax:   &budgetMax,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if request.Title != "Need Help" {
		t.Errorf("expected title 'Need Help', got %s", request.Title)
	}
	if request.RequesterID != requesterID {
		t.Errorf("expected requester ID %s, got %s", requesterID, request.RequesterID)
	}
	if request.Status != RequestStatusOpen {
		t.Errorf("expected status open, got %s", request.Status)
	}
}

func TestService_CreateRequest_MissingTitle(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	_, err := service.CreateRequest(context.Background(), uuid.New(), &CreateRequestRequest{
		RequestType: ListingTypeServices,
	})

	if err == nil {
		t.Error("expected error for missing title")
	}
}

func TestService_CreateRequest_MissingType(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	_, err := service.CreateRequest(context.Background(), uuid.New(), &CreateRequestRequest{
		Title: "Test",
	})

	if err == nil {
		t.Error("expected error for missing request type")
	}
}

func TestService_GetRequest(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	requesterID := uuid.New()
	request, _ := service.CreateRequest(context.Background(), requesterID, &CreateRequestRequest{
		Title:       "Test Request",
		RequestType: ListingTypeData,
	})

	retrieved, err := service.GetRequest(context.Background(), request.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.ID != request.ID {
		t.Errorf("expected ID %s, got %s", request.ID, retrieved.ID)
	}
}

func TestService_GetRequest_NotFound(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	_, err := service.GetRequest(context.Background(), uuid.New())
	if err != ErrRequestNotFound {
		t.Errorf("expected ErrRequestNotFound, got %v", err)
	}
}

func TestService_SearchRequests(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	requesterID := uuid.New()

	// Create some requests
	for i := 0; i < 3; i++ {
		service.CreateRequest(context.Background(), requesterID, &CreateRequestRequest{
			Title:       "Test Request",
			RequestType: ListingTypeServices,
		})
	}

	result, err := service.SearchRequests(context.Background(), SearchRequestsParams{
		RequesterID: &requesterID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("expected 3 requests, got %d", result.Total)
	}
}

func TestService_SubmitOffer(t *testing.T) {
	repo := newMockRepository()
	publisher := &mockPublisher{}
	service := NewService(repo, publisher)

	requesterID := uuid.New()
	offererID := uuid.New()

	request, _ := service.CreateRequest(context.Background(), requesterID, &CreateRequestRequest{
		Title:       "Need Help",
		RequestType: ListingTypeServices,
	})

	offer, err := service.SubmitOffer(context.Background(), offererID, request.ID, &CreateOfferRequest{
		PriceAmount:   100.0,
		Description:   "I can help",
		DeliveryTerms: "2 days",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if offer.OffererID != offererID {
		t.Errorf("expected offerer ID %s, got %s", offererID, offer.OffererID)
	}
	if offer.RequestID != request.ID {
		t.Errorf("expected request ID %s, got %s", request.ID, offer.RequestID)
	}
	if offer.Status != OfferStatusPending {
		t.Errorf("expected status pending, got %s", offer.Status)
	}
	if offer.PriceCurrency != "USD" {
		t.Errorf("expected currency USD, got %s", offer.PriceCurrency)
	}
}

func TestService_SubmitOffer_RequestNotFound(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	_, err := service.SubmitOffer(context.Background(), uuid.New(), uuid.New(), &CreateOfferRequest{
		PriceAmount: 100.0,
	})

	if err != ErrRequestNotFound {
		t.Errorf("expected ErrRequestNotFound, got %v", err)
	}
}

func TestService_SubmitOffer_SelfOffer(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	requesterID := uuid.New()
	request, _ := service.CreateRequest(context.Background(), requesterID, &CreateRequestRequest{
		Title:       "Need Help",
		RequestType: ListingTypeServices,
	})

	_, err := service.SubmitOffer(context.Background(), requesterID, request.ID, &CreateOfferRequest{
		PriceAmount: 100.0,
	})

	if err == nil {
		t.Error("expected error for self-offer")
	}
}

func TestService_SubmitOffer_InvalidPrice(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	requesterID := uuid.New()
	offererID := uuid.New()

	request, _ := service.CreateRequest(context.Background(), requesterID, &CreateRequestRequest{
		Title:       "Need Help",
		RequestType: ListingTypeServices,
	})

	tests := []float64{0, -50, -0.01}
	for _, price := range tests {
		_, err := service.SubmitOffer(context.Background(), offererID, request.ID, &CreateOfferRequest{
			PriceAmount: price,
		})
		if err == nil {
			t.Errorf("expected error for price %f", price)
		}
	}
}

func TestService_GetOffersByRequest(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	requesterID := uuid.New()
	request, _ := service.CreateRequest(context.Background(), requesterID, &CreateRequestRequest{
		Title:       "Need Help",
		RequestType: ListingTypeServices,
	})

	// Submit some offers
	for i := 0; i < 3; i++ {
		service.SubmitOffer(context.Background(), uuid.New(), request.ID, &CreateOfferRequest{
			PriceAmount: float64(100 * (i + 1)),
		})
	}

	offers, err := service.GetOffersByRequest(context.Background(), request.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(offers) != 3 {
		t.Errorf("expected 3 offers, got %d", len(offers))
	}
}

func TestService_AcceptOffer(t *testing.T) {
	repo := newMockRepository()
	publisher := &mockPublisher{}
	txCreator := newMockTransactionCreator()
	service := NewService(repo, publisher)
	service.SetTransactionCreator(txCreator)

	requesterID := uuid.New()
	offererID := uuid.New()

	request, _ := service.CreateRequest(context.Background(), requesterID, &CreateRequestRequest{
		Title:       "Need Help",
		RequestType: ListingTypeServices,
	})

	offer, _ := service.SubmitOffer(context.Background(), offererID, request.ID, &CreateOfferRequest{
		PriceAmount: 100.0,
	})

	accepted, err := service.AcceptOffer(context.Background(), requesterID, offer.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if accepted.Status != OfferStatusAccepted {
		t.Errorf("expected status accepted, got %s", accepted.Status)
	}

	// Verify transaction was created
	if len(txCreator.transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(txCreator.transactions))
	}
}

func TestService_AcceptOffer_NotRequester(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	requesterID := uuid.New()
	offererID := uuid.New()

	request, _ := service.CreateRequest(context.Background(), requesterID, &CreateRequestRequest{
		Title:       "Need Help",
		RequestType: ListingTypeServices,
	})

	offer, _ := service.SubmitOffer(context.Background(), offererID, request.ID, &CreateOfferRequest{
		PriceAmount: 100.0,
	})

	_, err := service.AcceptOffer(context.Background(), uuid.New(), offer.ID) // Different user
	if err == nil {
		t.Error("expected error for non-requester accepting")
	}
}

func TestService_AcceptOffer_NotPending(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	requesterID := uuid.New()
	offererID := uuid.New()

	request, _ := service.CreateRequest(context.Background(), requesterID, &CreateRequestRequest{
		Title:       "Need Help",
		RequestType: ListingTypeServices,
	})

	offer, _ := service.SubmitOffer(context.Background(), offererID, request.ID, &CreateOfferRequest{
		PriceAmount: 100.0,
	})

	// Accept once
	service.AcceptOffer(context.Background(), requesterID, offer.ID)

	// Try to accept again
	_, err := service.AcceptOffer(context.Background(), requesterID, offer.ID)
	if err == nil {
		t.Error("expected error for non-pending offer")
	}
}

func TestService_GetCategories(t *testing.T) {
	repo := newMockRepository()
	repo.categories = []Category{
		{ID: uuid.New(), Name: "Services", Slug: "services"},
		{ID: uuid.New(), Name: "Data", Slug: "data"},
	}
	service := NewService(repo, nil)

	categories, err := service.GetCategories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(categories) != 2 {
		t.Errorf("expected 2 categories, got %d", len(categories))
	}
}

func TestNewService(t *testing.T) {
	repo := newMockRepository()
	publisher := &mockPublisher{}
	service := NewService(repo, publisher)

	if service == nil {
		t.Fatal("expected service to be created")
	}
	if service.publisher != publisher {
		t.Error("publisher not set correctly")
	}
	if service.transactionCreator != nil {
		t.Error("transaction creator should be nil initially")
	}
}

func TestSetTransactionCreator(t *testing.T) {
	service := NewService(newMockRepository(), nil)
	txCreator := newMockTransactionCreator()

	service.SetTransactionCreator(txCreator)

	if service.transactionCreator == nil {
		t.Error("transaction creator should be set")
	}
}

func TestDefaultHelpers(t *testing.T) {
	// Test defaultString
	if defaultString("", "default") != "default" {
		t.Error("defaultString should return default for empty string")
	}
	if defaultString("value", "default") != "value" {
		t.Error("defaultString should return value when not empty")
	}

	// Test defaultInt
	if defaultInt(0, 10) != 10 {
		t.Error("defaultInt should return default for zero")
	}
	if defaultInt(5, 10) != 5 {
		t.Error("defaultInt should return value when not zero")
	}

	// Test defaultScope
	if defaultScope("", ScopeLocal) != ScopeLocal {
		t.Error("defaultScope should return default for empty")
	}
	if defaultScope(ScopeNational, ScopeLocal) != ScopeNational {
		t.Error("defaultScope should return value when not empty")
	}
}

func TestRepositoryErrors(t *testing.T) {
	if ErrListingNotFound.Error() != "listing not found" {
		t.Errorf("unexpected error message: %s", ErrListingNotFound.Error())
	}
	if ErrRequestNotFound.Error() != "request not found" {
		t.Errorf("unexpected error message: %s", ErrRequestNotFound.Error())
	}
	if ErrOfferNotFound.Error() != "offer not found" {
		t.Errorf("unexpected error message: %s", ErrOfferNotFound.Error())
	}
}
