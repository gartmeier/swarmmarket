package auction

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAuctionTypeValidation(t *testing.T) {
	tests := []struct {
		name        string
		auctionType string
		valid       bool
	}{
		{"english auction", "english", true},
		{"dutch auction", "dutch", true},
		{"sealed auction", "sealed", true},
		{"continuous auction", "continuous", true},
		{"invalid type", "invalid", false},
		{"empty type", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			auctionType := AuctionType(tc.auctionType)
			isValid := auctionType == AuctionTypeEnglish ||
				auctionType == AuctionTypeDutch ||
				auctionType == AuctionTypeSealed ||
				auctionType == AuctionTypeContinuous

			if isValid != tc.valid {
				t.Errorf("AuctionType(%q) validity = %v, want %v", tc.auctionType, isValid, tc.valid)
			}
		})
	}
}

func TestAuctionStatusValidation(t *testing.T) {
	tests := []struct {
		name   string
		status AuctionStatus
		valid  bool
	}{
		{"scheduled", AuctionStatusScheduled, true},
		{"active", AuctionStatusActive, true},
		{"ended", AuctionStatusEnded, true},
		{"cancelled", AuctionStatusCancelled, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.status == "" && tc.valid {
				t.Errorf("Expected status %q to be valid", tc.status)
			}
		})
	}
}

func TestBidStatusValidation(t *testing.T) {
	tests := []struct {
		name   string
		status BidStatus
		valid  bool
	}{
		{"active", BidStatusActive, true},
		{"outbid", BidStatusOutbid, true},
		{"winning", BidStatusWinning, true},
		{"won", BidStatusWon, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.status == "" && tc.valid {
				t.Errorf("Expected status %q to be valid", tc.status)
			}
		})
	}
}

func TestCreateAuctionRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateAuctionRequest
		wantErr bool
	}{
		{
			name: "valid english auction",
			req: CreateAuctionRequest{
				AuctionType:   "english",
				Title:         "Test Auction",
				StartingPrice: 100,
				EndsAt:        time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
		},
		{
			name: "valid dutch auction",
			req: CreateAuctionRequest{
				AuctionType:   "dutch",
				Title:         "Test Dutch Auction",
				StartingPrice: 500,
				EndsAt:        time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
		},
		{
			name: "missing title",
			req: CreateAuctionRequest{
				AuctionType:   "english",
				StartingPrice: 100,
				EndsAt:        time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "invalid auction type",
			req: CreateAuctionRequest{
				AuctionType:   "invalid",
				Title:         "Test Auction",
				StartingPrice: 100,
				EndsAt:        time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "zero starting price",
			req: CreateAuctionRequest{
				AuctionType:   "english",
				Title:         "Test Auction",
				StartingPrice: 0,
				EndsAt:        time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "negative starting price",
			req: CreateAuctionRequest{
				AuctionType:   "english",
				Title:         "Test Auction",
				StartingPrice: -100,
				EndsAt:        time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "end time in past",
			req: CreateAuctionRequest{
				AuctionType:   "english",
				Title:         "Test Auction",
				StartingPrice: 100,
				EndsAt:        time.Now().Add(-24 * time.Hour),
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hasError := false

			// Validate auction type
			auctionType := AuctionType(tc.req.AuctionType)
			switch auctionType {
			case AuctionTypeEnglish, AuctionTypeDutch, AuctionTypeSealed, AuctionTypeContinuous:
				// Valid
			default:
				hasError = true
			}

			// Validate title
			if tc.req.Title == "" {
				hasError = true
			}

			// Validate starting price
			if tc.req.StartingPrice <= 0 {
				hasError = true
			}

			// Validate end time
			if tc.req.EndsAt.Before(time.Now().UTC()) {
				hasError = true
			}

			if hasError != tc.wantErr {
				t.Errorf("CreateAuctionRequest validation error = %v, wantErr %v", hasError, tc.wantErr)
			}
		})
	}
}

func TestPlaceBidRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     PlaceBidRequest
		wantErr bool
	}{
		{
			name:    "valid bid",
			req:     PlaceBidRequest{Amount: 100},
			wantErr: false,
		},
		{
			name:    "zero amount",
			req:     PlaceBidRequest{Amount: 0},
			wantErr: true,
		},
		{
			name:    "negative amount",
			req:     PlaceBidRequest{Amount: -50},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hasError := tc.req.Amount <= 0

			if hasError != tc.wantErr {
				t.Errorf("PlaceBidRequest validation error = %v, wantErr %v", hasError, tc.wantErr)
			}
		})
	}
}

// mockPublisher implements EventPublisher for testing.
type mockPublisher struct {
	events []struct {
		eventType string
		payload   map[string]any
	}
}

func (m *mockPublisher) Publish(ctx context.Context, eventType string, payload map[string]any) error {
	m.events = append(m.events, struct {
		eventType string
		payload   map[string]any
	}{eventType, payload})
	return nil
}

// mockRepository implements RepositoryInterface for testing.
type mockRepository struct {
	auctions    map[uuid.UUID]*Auction
	bids        map[uuid.UUID][]*Bid
	createErr   error
	getErr      error
	searchErr   error
	createBidErr error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		auctions: make(map[uuid.UUID]*Auction),
		bids:     make(map[uuid.UUID][]*Bid),
	}
}

// Verify mockRepository implements RepositoryInterface
var _ RepositoryInterface = (*mockRepository)(nil)

func (m *mockRepository) CreateAuction(ctx context.Context, auction *Auction) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.auctions[auction.ID] = auction
	return nil
}

func (m *mockRepository) GetAuctionByID(ctx context.Context, id uuid.UUID) (*Auction, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	auction, ok := m.auctions[id]
	if !ok {
		return nil, ErrAuctionNotFound
	}
	// Count bids
	auction.BidCount = len(m.bids[id])
	return auction, nil
}

func (m *mockRepository) SearchAuctions(ctx context.Context, params SearchAuctionsParams) (*AuctionListResult, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	var auctions []*Auction
	for _, a := range m.auctions {
		if params.SellerID != nil && a.SellerID != *params.SellerID {
			continue
		}
		if params.AuctionType != nil && a.AuctionType != *params.AuctionType {
			continue
		}
		if params.Status != nil && a.Status != *params.Status {
			continue
		}
		auctions = append(auctions, a)
	}
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	return &AuctionListResult{
		Auctions: auctions,
		Total:    len(auctions),
		Limit:    limit,
		Offset:   params.Offset,
	}, nil
}

func (m *mockRepository) UpdateAuctionPrice(ctx context.Context, auctionID uuid.UUID, price float64) error {
	auction, ok := m.auctions[auctionID]
	if !ok {
		return ErrAuctionNotFound
	}
	auction.CurrentPrice = &price
	return nil
}

func (m *mockRepository) UpdateAuctionStatus(ctx context.Context, auctionID uuid.UUID, status AuctionStatus) error {
	auction, ok := m.auctions[auctionID]
	if !ok {
		return ErrAuctionNotFound
	}
	auction.Status = status
	return nil
}

func (m *mockRepository) SetAuctionWinner(ctx context.Context, auctionID, winningBidID, winnerID uuid.UUID) error {
	auction, ok := m.auctions[auctionID]
	if !ok {
		return ErrAuctionNotFound
	}
	auction.WinningBidID = &winningBidID
	auction.WinnerID = &winnerID
	auction.Status = AuctionStatusEnded
	return nil
}

func (m *mockRepository) ExtendAuction(ctx context.Context, auctionID uuid.UUID, newEndTime time.Time) error {
	auction, ok := m.auctions[auctionID]
	if !ok {
		return ErrAuctionNotFound
	}
	auction.EndsAt = newEndTime
	return nil
}

func (m *mockRepository) CreateBid(ctx context.Context, bid *Bid) error {
	if m.createBidErr != nil {
		return m.createBidErr
	}
	m.bids[bid.AuctionID] = append(m.bids[bid.AuctionID], bid)
	return nil
}

func (m *mockRepository) GetBidsByAuctionID(ctx context.Context, auctionID uuid.UUID) ([]*Bid, error) {
	// Return copies to avoid service modifications affecting stored data
	var result []*Bid
	for _, b := range m.bids[auctionID] {
		copy := *b
		result = append(result, &copy)
	}
	return result, nil
}

func (m *mockRepository) GetHighestBid(ctx context.Context, auctionID uuid.UUID) (*Bid, error) {
	bids := m.bids[auctionID]
	if len(bids) == 0 {
		return nil, nil
	}
	var highest *Bid
	for _, b := range bids {
		if b.Status == BidStatusActive || b.Status == BidStatusWinning {
			if highest == nil || b.Amount > highest.Amount {
				highest = b
			}
		}
	}
	return highest, nil
}

func (m *mockRepository) UpdateBidStatus(ctx context.Context, bidID uuid.UUID, status BidStatus) error {
	for _, bids := range m.bids {
		for _, b := range bids {
			if b.ID == bidID {
				b.Status = status
				return nil
			}
		}
	}
	return ErrBidNotFound
}

func (m *mockRepository) MarkPreviousBidsOutbid(ctx context.Context, auctionID uuid.UUID, exceptBidID uuid.UUID) error {
	for _, b := range m.bids[auctionID] {
		if b.ID != exceptBidID && b.Status == BidStatusActive {
			b.Status = BidStatusOutbid
		}
	}
	return nil
}

func TestNewService(t *testing.T) {
	publisher := &mockPublisher{}
	service := NewService(nil, publisher)

	if service == nil {
		t.Fatal("expected service to be created")
	}
	if service.publisher != publisher {
		t.Error("publisher not set correctly")
	}
}

func TestServiceErrors(t *testing.T) {
	tests := []struct {
		err      error
		expected string
	}{
		{ErrAuctionNotActive, "auction is not active"},
		{ErrAuctionEnded, "auction has ended"},
		{ErrBidTooLow, "bid amount is too low"},
		{ErrCannotBidOnOwnAuction, "cannot bid on your own auction"},
		{ErrNotAuthorized, "not authorized to perform this action"},
		{ErrInvalidAuctionType, "invalid auction type"},
	}

	for _, tt := range tests {
		if tt.err.Error() != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, tt.err.Error())
		}
	}
}

func TestAuction_AllFields(t *testing.T) {
	now := time.Now()
	listingID := uuid.New()
	winningBidID := uuid.New()
	winnerID := uuid.New()
	currentPrice := 150.0
	reservePrice := 100.0
	buyNowPrice := 500.0
	minIncrement := 10.0
	priceDecrement := 5.0
	decrementInterval := 300

	auction := &Auction{
		ID:                    uuid.New(),
		ListingID:             &listingID,
		SellerID:              uuid.New(),
		AuctionType:           AuctionTypeEnglish,
		Title:                 "Test Auction",
		Description:           "A test auction",
		StartingPrice:         100.0,
		CurrentPrice:          &currentPrice,
		ReservePrice:          &reservePrice,
		BuyNowPrice:           &buyNowPrice,
		Currency:              "USD",
		MinIncrement:          &minIncrement,
		PriceDecrement:        &priceDecrement,
		DecrementIntervalSecs: &decrementInterval,
		Status:                AuctionStatusActive,
		StartsAt:              now,
		EndsAt:                now.Add(24 * time.Hour),
		ExtensionSeconds:      60,
		WinningBidID:          &winningBidID,
		WinnerID:              &winnerID,
		BidCount:              5,
		Metadata:              map[string]any{"category": "data"},
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if auction.Title != "Test Auction" {
		t.Errorf("expected title 'Test Auction', got %s", auction.Title)
	}
	if auction.AuctionType != AuctionTypeEnglish {
		t.Error("auction type not set correctly")
	}
	if *auction.CurrentPrice != 150.0 {
		t.Error("current price not set correctly")
	}
	if auction.BidCount != 5 {
		t.Error("bid count not set correctly")
	}
	if *auction.ListingID != listingID {
		t.Error("listing ID not set correctly")
	}
	if *auction.WinnerID != winnerID {
		t.Error("winner ID not set correctly")
	}
}

func TestBid_AllFields(t *testing.T) {
	now := time.Now()
	bid := &Bid{
		ID:        uuid.New(),
		AuctionID: uuid.New(),
		BidderID:  uuid.New(),
		Amount:    200.0,
		Currency:  "USD",
		IsSealed:  true,
		Status:    BidStatusWinning,
		Metadata:  map[string]any{"auto": true},
		CreatedAt: now,
	}

	if bid.Amount != 200.0 {
		t.Errorf("expected amount 200.0, got %f", bid.Amount)
	}
	if bid.Status != BidStatusWinning {
		t.Error("status not set correctly")
	}
	if !bid.IsSealed {
		t.Error("IsSealed should be true")
	}
	if bid.Metadata["auto"] != true {
		t.Error("metadata not set correctly")
	}
}

func TestAuctionListResult(t *testing.T) {
	result := &AuctionListResult{
		Auctions: []*Auction{{Title: "Auction 1"}, {Title: "Auction 2"}},
		Total:    2,
		Limit:    20,
		Offset:   0,
	}

	if len(result.Auctions) != 2 {
		t.Errorf("expected 2 auctions, got %d", len(result.Auctions))
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if result.Limit != 20 {
		t.Errorf("expected limit 20, got %d", result.Limit)
	}
}

func TestSearchAuctionsParams(t *testing.T) {
	sellerID := uuid.New()
	auctionType := AuctionTypeEnglish
	status := AuctionStatusActive

	params := SearchAuctionsParams{
		SellerID:    &sellerID,
		AuctionType: &auctionType,
		Status:      &status,
		Query:       "test",
		Limit:       20,
		Offset:      10,
	}

	if *params.SellerID != sellerID {
		t.Error("seller ID not set correctly")
	}
	if *params.AuctionType != AuctionTypeEnglish {
		t.Error("auction type not set correctly")
	}
	if *params.Status != AuctionStatusActive {
		t.Error("status not set correctly")
	}
	if params.Query != "test" {
		t.Errorf("expected query 'test', got %s", params.Query)
	}
	if params.Offset != 10 {
		t.Errorf("expected offset 10, got %d", params.Offset)
	}
}

func TestDutchAuctionFields(t *testing.T) {
	priceDecrement := 5.0
	decrementInterval := 300

	auction := &Auction{
		ID:                    uuid.New(),
		AuctionType:           AuctionTypeDutch,
		Title:                 "Dutch Auction",
		StartingPrice:         1000.0,
		PriceDecrement:        &priceDecrement,
		DecrementIntervalSecs: &decrementInterval,
		Status:                AuctionStatusActive,
	}

	if auction.AuctionType != AuctionTypeDutch {
		t.Error("auction type should be dutch")
	}
	if *auction.PriceDecrement != 5.0 {
		t.Error("price decrement not set correctly")
	}
	if *auction.DecrementIntervalSecs != 300 {
		t.Error("decrement interval not set correctly")
	}
}

func TestSealedBid(t *testing.T) {
	bid := &Bid{
		ID:        uuid.New(),
		AuctionID: uuid.New(),
		BidderID:  uuid.New(),
		Amount:    500.0,
		IsSealed:  true,
		Status:    BidStatusActive,
	}

	if !bid.IsSealed {
		t.Error("bid should be sealed")
	}
}

func TestPublishEvent(t *testing.T) {
	publisher := &mockPublisher{}
	service := &Service{publisher: publisher}

	ctx := context.Background()
	service.publishEvent(ctx, "test.event", map[string]any{"key": "value"})

	// Give the goroutine time to execute
	time.Sleep(10 * time.Millisecond)

	if len(publisher.events) != 1 {
		t.Errorf("expected 1 event, got %d", len(publisher.events))
	}
	if publisher.events[0].eventType != "test.event" {
		t.Errorf("expected event type 'test.event', got %s", publisher.events[0].eventType)
	}
}

func TestPublishEventNilPublisher(t *testing.T) {
	service := &Service{publisher: nil}

	// Should not panic with nil publisher
	ctx := context.Background()
	service.publishEvent(ctx, "test.event", map[string]any{"key": "value"})
}

func TestCreateAuctionRequestAllFields(t *testing.T) {
	listingID := uuid.New()
	reservePrice := 100.0
	buyNowPrice := 500.0
	minIncrement := 10.0
	extensionSeconds := 120
	startsAt := time.Now()
	endsAt := time.Now().Add(24 * time.Hour)
	priceDecrement := 5.0
	decrementInterval := 300

	req := &CreateAuctionRequest{
		ListingID:             &listingID,
		AuctionType:           "english",
		Title:                 "Test Auction",
		Description:           "A test auction description",
		StartingPrice:         100.0,
		ReservePrice:          &reservePrice,
		BuyNowPrice:           &buyNowPrice,
		Currency:              "USD",
		MinIncrement:          &minIncrement,
		PriceDecrement:        &priceDecrement,
		DecrementIntervalSecs: &decrementInterval,
		StartsAt:              &startsAt,
		EndsAt:                endsAt,
		ExtensionSeconds:      &extensionSeconds,
	}

	if *req.ListingID != listingID {
		t.Error("listing ID not set correctly")
	}
	if req.Description != "A test auction description" {
		t.Error("description not set correctly")
	}
	if *req.ReservePrice != 100.0 {
		t.Error("reserve price not set correctly")
	}
	if *req.BuyNowPrice != 500.0 {
		t.Error("buy now price not set correctly")
	}
	if *req.PriceDecrement != 5.0 {
		t.Error("price decrement not set correctly")
	}
}

// --- Service Method Tests ---

func TestService_CreateAuction(t *testing.T) {
	repo := newMockRepository()
	publisher := &mockPublisher{}
	service := NewService(repo, publisher)

	sellerID := uuid.New()

	auction, err := service.CreateAuction(context.Background(), sellerID, &CreateAuctionRequest{
		AuctionType:   "english",
		Title:         "Test Auction",
		Description:   "A test auction",
		StartingPrice: 100.0,
		EndsAt:        time.Now().Add(24 * time.Hour),
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if auction.Title != "Test Auction" {
		t.Errorf("expected title 'Test Auction', got %s", auction.Title)
	}
	if auction.SellerID != sellerID {
		t.Errorf("expected seller ID %s, got %s", sellerID, auction.SellerID)
	}
	if auction.Status != AuctionStatusActive {
		t.Errorf("expected status active, got %s", auction.Status)
	}
	if auction.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", auction.Currency)
	}
	if auction.ExtensionSeconds != 60 {
		t.Errorf("expected extension 60, got %d", auction.ExtensionSeconds)
	}
}

func TestService_CreateAuction_InvalidType(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	_, err := service.CreateAuction(context.Background(), uuid.New(), &CreateAuctionRequest{
		AuctionType:   "invalid",
		Title:         "Test",
		StartingPrice: 100.0,
		EndsAt:        time.Now().Add(24 * time.Hour),
	})

	if err != ErrInvalidAuctionType {
		t.Errorf("expected ErrInvalidAuctionType, got %v", err)
	}
}

func TestService_CreateAuction_InvalidPrice(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	tests := []float64{0, -100}
	for _, price := range tests {
		_, err := service.CreateAuction(context.Background(), uuid.New(), &CreateAuctionRequest{
			AuctionType:   "english",
			Title:         "Test",
			StartingPrice: price,
			EndsAt:        time.Now().Add(24 * time.Hour),
		})
		if err == nil {
			t.Errorf("expected error for price %f", price)
		}
	}
}

func TestService_CreateAuction_InvalidEndTime(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	_, err := service.CreateAuction(context.Background(), uuid.New(), &CreateAuctionRequest{
		AuctionType:   "english",
		Title:         "Test",
		StartingPrice: 100.0,
		EndsAt:        time.Now().Add(-24 * time.Hour), // In the past
	})

	if err == nil {
		t.Error("expected error for past end time")
	}
}

func TestService_CreateAuction_Scheduled(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	startsAt := time.Now().Add(1 * time.Hour)

	auction, err := service.CreateAuction(context.Background(), uuid.New(), &CreateAuctionRequest{
		AuctionType:   "english",
		Title:         "Scheduled Auction",
		StartingPrice: 100.0,
		StartsAt:      &startsAt,
		EndsAt:        time.Now().Add(24 * time.Hour),
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if auction.Status != AuctionStatusScheduled {
		t.Errorf("expected status scheduled, got %s", auction.Status)
	}
}

func TestService_GetAuction(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	sellerID := uuid.New()
	auction, _ := service.CreateAuction(context.Background(), sellerID, &CreateAuctionRequest{
		AuctionType:   "english",
		Title:         "Test Auction",
		StartingPrice: 100.0,
		EndsAt:        time.Now().Add(24 * time.Hour),
	})

	retrieved, err := service.GetAuction(context.Background(), auction.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.ID != auction.ID {
		t.Errorf("expected ID %s, got %s", auction.ID, retrieved.ID)
	}
}

func TestService_GetAuction_NotFound(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	_, err := service.GetAuction(context.Background(), uuid.New())
	if err != ErrAuctionNotFound {
		t.Errorf("expected ErrAuctionNotFound, got %v", err)
	}
}

func TestService_SearchAuctions(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	sellerID := uuid.New()

	// Create some auctions
	for i := 0; i < 3; i++ {
		service.CreateAuction(context.Background(), sellerID, &CreateAuctionRequest{
			AuctionType:   "english",
			Title:         "Test Auction",
			StartingPrice: float64(100 * (i + 1)),
			EndsAt:        time.Now().Add(24 * time.Hour),
		})
	}

	result, err := service.SearchAuctions(context.Background(), SearchAuctionsParams{
		SellerID: &sellerID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("expected 3 auctions, got %d", result.Total)
	}
}

func TestService_PlaceBid_English(t *testing.T) {
	repo := newMockRepository()
	publisher := &mockPublisher{}
	service := NewService(repo, publisher)

	sellerID := uuid.New()
	bidderID := uuid.New()

	auction, _ := service.CreateAuction(context.Background(), sellerID, &CreateAuctionRequest{
		AuctionType:   "english",
		Title:         "English Auction",
		StartingPrice: 100.0,
		EndsAt:        time.Now().Add(24 * time.Hour),
	})

	bid, err := service.PlaceBid(context.Background(), auction.ID, bidderID, &PlaceBidRequest{
		Amount: 100.0,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bid.Amount != 100.0 {
		t.Errorf("expected amount 100.0, got %f", bid.Amount)
	}
	if bid.BidderID != bidderID {
		t.Errorf("expected bidder ID %s, got %s", bidderID, bid.BidderID)
	}
	if bid.Status != BidStatusActive {
		t.Errorf("expected status active, got %s", bid.Status)
	}
}

func TestService_PlaceBid_TooLow(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	sellerID := uuid.New()
	bidderID := uuid.New()

	auction, _ := service.CreateAuction(context.Background(), sellerID, &CreateAuctionRequest{
		AuctionType:   "english",
		Title:         "English Auction",
		StartingPrice: 100.0,
		EndsAt:        time.Now().Add(24 * time.Hour),
	})

	// Place first bid
	service.PlaceBid(context.Background(), auction.ID, bidderID, &PlaceBidRequest{
		Amount: 100.0,
	})

	// Try to place lower bid
	_, err := service.PlaceBid(context.Background(), auction.ID, uuid.New(), &PlaceBidRequest{
		Amount: 99.0,
	})

	if err != ErrBidTooLow {
		t.Errorf("expected ErrBidTooLow, got %v", err)
	}
}

func TestService_PlaceBid_OwnAuction(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	sellerID := uuid.New()

	auction, _ := service.CreateAuction(context.Background(), sellerID, &CreateAuctionRequest{
		AuctionType:   "english",
		Title:         "English Auction",
		StartingPrice: 100.0,
		EndsAt:        time.Now().Add(24 * time.Hour),
	})

	_, err := service.PlaceBid(context.Background(), auction.ID, sellerID, &PlaceBidRequest{
		Amount: 100.0,
	})

	if err != ErrCannotBidOnOwnAuction {
		t.Errorf("expected ErrCannotBidOnOwnAuction, got %v", err)
	}
}

func TestService_PlaceBid_NotActive(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	sellerID := uuid.New()

	auction, _ := service.CreateAuction(context.Background(), sellerID, &CreateAuctionRequest{
		AuctionType:   "english",
		Title:         "English Auction",
		StartingPrice: 100.0,
		EndsAt:        time.Now().Add(24 * time.Hour),
	})

	// Set to ended
	repo.UpdateAuctionStatus(context.Background(), auction.ID, AuctionStatusEnded)

	_, err := service.PlaceBid(context.Background(), auction.ID, uuid.New(), &PlaceBidRequest{
		Amount: 100.0,
	})

	if err != ErrAuctionNotActive {
		t.Errorf("expected ErrAuctionNotActive, got %v", err)
	}
}

func TestService_PlaceBid_Dutch(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	sellerID := uuid.New()
	bidderID := uuid.New()

	auction, _ := service.CreateAuction(context.Background(), sellerID, &CreateAuctionRequest{
		AuctionType:   "dutch",
		Title:         "Dutch Auction",
		StartingPrice: 500.0,
		EndsAt:        time.Now().Add(24 * time.Hour),
	})

	bid, err := service.PlaceBid(context.Background(), auction.ID, bidderID, &PlaceBidRequest{
		Amount: 500.0,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Dutch auction should end immediately
	updated, _ := service.GetAuction(context.Background(), auction.ID)
	if updated.Status != AuctionStatusEnded {
		t.Errorf("expected auction to be ended, got %s", updated.Status)
	}
	if updated.WinnerID == nil || *updated.WinnerID != bidderID {
		t.Error("winner should be set")
	}
	if bid.Status != BidStatusWon {
		t.Errorf("expected bid status won, got %s", bid.Status)
	}
}

func TestService_PlaceBid_Sealed(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	sellerID := uuid.New()

	auction, _ := service.CreateAuction(context.Background(), sellerID, &CreateAuctionRequest{
		AuctionType:   "sealed",
		Title:         "Sealed Auction",
		StartingPrice: 100.0,
		EndsAt:        time.Now().Add(24 * time.Hour),
	})

	bid, err := service.PlaceBid(context.Background(), auction.ID, uuid.New(), &PlaceBidRequest{
		Amount: 200.0,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bid.IsSealed {
		t.Error("bid should be sealed")
	}
}

func TestService_GetBids(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	sellerID := uuid.New()

	auction, _ := service.CreateAuction(context.Background(), sellerID, &CreateAuctionRequest{
		AuctionType:   "english",
		Title:         "English Auction",
		StartingPrice: 100.0,
		EndsAt:        time.Now().Add(24 * time.Hour),
	})

	// Place some bids
	for i := 0; i < 3; i++ {
		service.PlaceBid(context.Background(), auction.ID, uuid.New(), &PlaceBidRequest{
			Amount: float64(100 + i*10),
		})
	}

	bids, err := service.GetBids(context.Background(), auction.ID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bids) != 3 {
		t.Errorf("expected 3 bids, got %d", len(bids))
	}
}

func TestService_GetBids_SealedHidden(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	sellerID := uuid.New()
	bidderID := uuid.New()

	auction, _ := service.CreateAuction(context.Background(), sellerID, &CreateAuctionRequest{
		AuctionType:   "sealed",
		Title:         "Sealed Auction",
		StartingPrice: 100.0,
		EndsAt:        time.Now().Add(24 * time.Hour),
	})

	// Place a bid
	service.PlaceBid(context.Background(), auction.ID, bidderID, &PlaceBidRequest{
		Amount: 200.0,
	})

	// Get bids as non-bidder - amounts should be hidden
	bids, _ := service.GetBids(context.Background(), auction.ID, nil)
	if bids[0].Amount != 0 {
		t.Error("bid amount should be hidden for non-bidder")
	}

	// Get bids as bidder - amounts should be visible
	bids, _ = service.GetBids(context.Background(), auction.ID, &bidderID)
	if bids[0].Amount != 200.0 {
		t.Error("bid amount should be visible for bidder")
	}
}

func TestService_EndAuction(t *testing.T) {
	repo := newMockRepository()
	publisher := &mockPublisher{}
	service := NewService(repo, publisher)

	sellerID := uuid.New()
	bidderID := uuid.New()

	auction, _ := service.CreateAuction(context.Background(), sellerID, &CreateAuctionRequest{
		AuctionType:   "english",
		Title:         "English Auction",
		StartingPrice: 100.0,
		EndsAt:        time.Now().Add(24 * time.Hour),
	})

	// Place a bid
	service.PlaceBid(context.Background(), auction.ID, bidderID, &PlaceBidRequest{
		Amount: 150.0,
	})

	// End auction
	ended, err := service.EndAuction(context.Background(), auction.ID, sellerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ended.Status != AuctionStatusEnded {
		t.Errorf("expected status ended, got %s", ended.Status)
	}
	if ended.WinnerID == nil || *ended.WinnerID != bidderID {
		t.Error("winner should be set")
	}
}

func TestService_EndAuction_NotAuthorized(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, nil)

	sellerID := uuid.New()

	auction, _ := service.CreateAuction(context.Background(), sellerID, &CreateAuctionRequest{
		AuctionType:   "english",
		Title:         "English Auction",
		StartingPrice: 100.0,
		EndsAt:        time.Now().Add(24 * time.Hour),
	})

	_, err := service.EndAuction(context.Background(), auction.ID, uuid.New()) // Different user
	if err != ErrNotAuthorized {
		t.Errorf("expected ErrNotAuthorized, got %v", err)
	}
}

func TestService_EndAuction_NoBids(t *testing.T) {
	repo := newMockRepository()
	publisher := &mockPublisher{}
	service := NewService(repo, publisher)

	sellerID := uuid.New()

	auction, _ := service.CreateAuction(context.Background(), sellerID, &CreateAuctionRequest{
		AuctionType:   "english",
		Title:         "English Auction",
		StartingPrice: 100.0,
		EndsAt:        time.Now().Add(24 * time.Hour),
	})

	// End with no bids
	ended, err := service.EndAuction(context.Background(), auction.ID, sellerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ended.Status != AuctionStatusEnded {
		t.Errorf("expected status ended, got %s", ended.Status)
	}
	if ended.WinnerID != nil {
		t.Error("winner should be nil with no bids")
	}
}

func TestService_EndAuction_ReserveNotMet(t *testing.T) {
	repo := newMockRepository()
	publisher := &mockPublisher{}
	service := NewService(repo, publisher)

	sellerID := uuid.New()
	reservePrice := 200.0

	auction, _ := service.CreateAuction(context.Background(), sellerID, &CreateAuctionRequest{
		AuctionType:   "english",
		Title:         "English Auction",
		StartingPrice: 100.0,
		ReservePrice:  &reservePrice,
		EndsAt:        time.Now().Add(24 * time.Hour),
	})

	// Place bid below reserve
	service.PlaceBid(context.Background(), auction.ID, uuid.New(), &PlaceBidRequest{
		Amount: 150.0,
	})

	// End auction - reserve not met
	ended, err := service.EndAuction(context.Background(), auction.ID, sellerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ended.WinnerID != nil {
		t.Error("winner should be nil when reserve not met")
	}
}

func TestRepositoryErrors(t *testing.T) {
	if ErrAuctionNotFound.Error() != "auction not found" {
		t.Errorf("unexpected error message: %s", ErrAuctionNotFound.Error())
	}
	if ErrBidNotFound.Error() != "bid not found" {
		t.Errorf("unexpected error message: %s", ErrBidNotFound.Error())
	}
}
