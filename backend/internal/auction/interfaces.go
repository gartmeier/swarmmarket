package auction

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RepositoryInterface defines the contract for auction data persistence.
// This interface enables mock implementations for testing.
type RepositoryInterface interface {
	// Auction Operations
	CreateAuction(ctx context.Context, auction *Auction) error
	GetAuctionByID(ctx context.Context, id uuid.UUID) (*Auction, error)
	SearchAuctions(ctx context.Context, params SearchAuctionsParams) (*AuctionListResult, error)
	UpdateAuctionPrice(ctx context.Context, auctionID uuid.UUID, price float64) error
	UpdateAuctionStatus(ctx context.Context, auctionID uuid.UUID, status AuctionStatus) error
	SetAuctionWinner(ctx context.Context, auctionID, winningBidID, winnerID uuid.UUID) error
	ExtendAuction(ctx context.Context, auctionID uuid.UUID, newEndTime time.Time) error

	// Bid Operations
	CreateBid(ctx context.Context, bid *Bid) error
	GetBidsByAuctionID(ctx context.Context, auctionID uuid.UUID) ([]*Bid, error)
	GetHighestBid(ctx context.Context, auctionID uuid.UUID) (*Bid, error)
	UpdateBidStatus(ctx context.Context, bidID uuid.UUID, status BidStatus) error
	MarkPreviousBidsOutbid(ctx context.Context, auctionID uuid.UUID, exceptBidID uuid.UUID) error
}

// Verify that Repository implements RepositoryInterface
var _ RepositoryInterface = (*Repository)(nil)
