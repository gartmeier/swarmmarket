package api

import (
	"context"

	"github.com/google/uuid"
)

// ListingOwnerChecker interface for checking listing ownership.
type ListingOwnerChecker interface {
	IsListingOwner(ctx context.Context, listingID, agentID uuid.UUID) (bool, error)
}

// RequestOwnerChecker interface for checking request ownership.
type RequestOwnerChecker interface {
	IsRequestOwner(ctx context.Context, requestID, agentID uuid.UUID) (bool, error)
}

// AuctionOwnerChecker interface for checking auction ownership.
type AuctionOwnerChecker interface {
	IsAuctionOwner(ctx context.Context, auctionID, agentID uuid.UUID) (bool, error)
}

// CombinedOwnershipChecker combines the ownership checkers.
type CombinedOwnershipChecker struct {
	ListingChecker ListingOwnerChecker
	RequestChecker RequestOwnerChecker
	AuctionChecker AuctionOwnerChecker
}

// IsListingOwner delegates to the listing checker.
func (c *CombinedOwnershipChecker) IsListingOwner(ctx context.Context, listingID, agentID uuid.UUID) (bool, error) {
	if c.ListingChecker == nil {
		return false, nil
	}
	return c.ListingChecker.IsListingOwner(ctx, listingID, agentID)
}

// IsRequestOwner delegates to the request checker.
func (c *CombinedOwnershipChecker) IsRequestOwner(ctx context.Context, requestID, agentID uuid.UUID) (bool, error) {
	if c.RequestChecker == nil {
		return false, nil
	}
	return c.RequestChecker.IsRequestOwner(ctx, requestID, agentID)
}

// IsAuctionOwner delegates to the auction checker.
func (c *CombinedOwnershipChecker) IsAuctionOwner(ctx context.Context, auctionID, agentID uuid.UUID) (bool, error) {
	if c.AuctionChecker == nil {
		return false, nil
	}
	return c.AuctionChecker.IsAuctionOwner(ctx, auctionID, agentID)
}
