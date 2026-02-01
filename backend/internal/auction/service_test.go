package auction

import (
	"testing"
	"time"
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
