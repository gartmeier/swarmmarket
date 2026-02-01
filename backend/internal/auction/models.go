package auction

import (
	"time"

	"github.com/google/uuid"
)

// AuctionType represents the type of auction.
type AuctionType string

const (
	AuctionTypeEnglish    AuctionType = "english"    // Ascending price, highest bidder wins
	AuctionTypeDutch      AuctionType = "dutch"      // Descending price, first bidder wins
	AuctionTypeSealed     AuctionType = "sealed"     // Sealed-bid, highest bid wins
	AuctionTypeContinuous AuctionType = "continuous" // Ongoing, like a limit order book
)

// AuctionStatus represents the status of an auction.
type AuctionStatus string

const (
	AuctionStatusScheduled AuctionStatus = "scheduled"
	AuctionStatusActive    AuctionStatus = "active"
	AuctionStatusEnded     AuctionStatus = "ended"
	AuctionStatusCancelled AuctionStatus = "cancelled"
)

// BidStatus represents the status of a bid.
type BidStatus string

const (
	BidStatusActive  BidStatus = "active"
	BidStatusOutbid  BidStatus = "outbid"
	BidStatusWinning BidStatus = "winning"
	BidStatusWon     BidStatus = "won"
)

// Auction represents an auction.
type Auction struct {
	ID                      uuid.UUID      `json:"id"`
	ListingID               *uuid.UUID     `json:"listing_id,omitempty"`
	SellerID                uuid.UUID      `json:"seller_id"`
	AuctionType             AuctionType    `json:"auction_type"`
	Title                   string         `json:"title"`
	Description             string         `json:"description,omitempty"`
	StartingPrice           float64        `json:"starting_price"`
	CurrentPrice            *float64       `json:"current_price,omitempty"`
	ReservePrice            *float64       `json:"reserve_price,omitempty"`
	BuyNowPrice             *float64       `json:"buy_now_price,omitempty"`
	Currency                string         `json:"currency"`
	MinIncrement            *float64       `json:"min_increment,omitempty"`             // For English auctions
	PriceDecrement          *float64       `json:"price_decrement,omitempty"`           // For Dutch auctions
	DecrementIntervalSecs   *int           `json:"decrement_interval_seconds,omitempty"` // For Dutch auctions
	Status                  AuctionStatus  `json:"status"`
	StartsAt                time.Time      `json:"starts_at"`
	EndsAt                  time.Time      `json:"ends_at"`
	ExtensionSeconds        int            `json:"extension_seconds"` // Anti-sniping
	WinningBidID            *uuid.UUID     `json:"winning_bid_id,omitempty"`
	WinnerID                *uuid.UUID     `json:"winner_id,omitempty"`
	BidCount                int            `json:"bid_count"`
	Metadata                map[string]any `json:"metadata,omitempty"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
}

// Bid represents a bid on an auction.
type Bid struct {
	ID        uuid.UUID      `json:"id"`
	AuctionID uuid.UUID      `json:"auction_id"`
	BidderID  uuid.UUID      `json:"bidder_id"`
	Amount    float64        `json:"amount"`
	Currency  string         `json:"currency"`
	IsSealed  bool           `json:"is_sealed"`
	Status    BidStatus      `json:"status"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// CreateAuctionRequest is the request for creating an auction.
type CreateAuctionRequest struct {
	ListingID             *uuid.UUID `json:"listing_id,omitempty"`
	AuctionType           string     `json:"auction_type"`
	Title                 string     `json:"title"`
	Description           string     `json:"description,omitempty"`
	StartingPrice         float64    `json:"starting_price"`
	ReservePrice          *float64   `json:"reserve_price,omitempty"`
	BuyNowPrice           *float64   `json:"buy_now_price,omitempty"`
	Currency              string     `json:"currency,omitempty"`
	MinIncrement          *float64   `json:"min_increment,omitempty"`
	PriceDecrement        *float64   `json:"price_decrement,omitempty"`
	DecrementIntervalSecs *int       `json:"decrement_interval_seconds,omitempty"`
	StartsAt              *time.Time `json:"starts_at,omitempty"`
	EndsAt                time.Time  `json:"ends_at"`
	ExtensionSeconds      *int       `json:"extension_seconds,omitempty"`
}

// PlaceBidRequest is the request for placing a bid.
type PlaceBidRequest struct {
	Amount float64 `json:"amount"`
}

// AuctionListResult is the result of listing auctions.
type AuctionListResult struct {
	Auctions []*Auction `json:"auctions"`
	Total    int        `json:"total"`
	Limit    int        `json:"limit"`
	Offset   int        `json:"offset"`
}

// SearchAuctionsParams are the parameters for searching auctions.
type SearchAuctionsParams struct {
	SellerID    *uuid.UUID
	AuctionType *AuctionType
	Status      *AuctionStatus
	Query       string
	Limit       int
	Offset      int
}
