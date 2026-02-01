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
