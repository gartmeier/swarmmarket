package auction

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAuctionNotFound = errors.New("auction not found")
	ErrBidNotFound     = errors.New("bid not found")
)

// Repository handles auction persistence.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new auction repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// CreateAuction creates a new auction.
func (r *Repository) CreateAuction(ctx context.Context, auction *Auction) error {
	query := `
		INSERT INTO auctions (
			id, listing_id, seller_id, auction_type, title, description,
			starting_price, current_price, reserve_price, buy_now_price, price_currency,
			min_increment, price_decrement, decrement_interval_seconds,
			status, starts_at, ends_at, extension_seconds, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14,
			$15, $16, $17, $18, $19, $20, $21
		)
	`

	_, err := r.pool.Exec(ctx, query,
		auction.ID,
		auction.ListingID,
		auction.SellerID,
		auction.AuctionType,
		auction.Title,
		auction.Description,
		auction.StartingPrice,
		auction.CurrentPrice,
		auction.ReservePrice,
		auction.BuyNowPrice,
		auction.Currency,
		auction.MinIncrement,
		auction.PriceDecrement,
		auction.DecrementIntervalSecs,
		auction.Status,
		auction.StartsAt,
		auction.EndsAt,
		auction.ExtensionSeconds,
		auction.Metadata,
		auction.CreatedAt,
		auction.UpdatedAt,
	)

	return err
}

// GetAuctionByID retrieves an auction by ID.
func (r *Repository) GetAuctionByID(ctx context.Context, id uuid.UUID) (*Auction, error) {
	query := `
		SELECT
			a.id, a.listing_id, a.seller_id, a.auction_type, a.title, a.description,
			a.starting_price, a.current_price, a.reserve_price, a.buy_now_price, a.price_currency,
			a.min_increment, a.price_decrement, a.decrement_interval_seconds,
			a.status, a.starts_at, a.ends_at, a.extension_seconds,
			a.winning_bid_id, a.winner_id, a.metadata, a.created_at, a.updated_at,
			(SELECT COUNT(*) FROM bids WHERE auction_id = a.id) as bid_count
		FROM auctions a
		WHERE a.id = $1
	`

	var auction Auction
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&auction.ID,
		&auction.ListingID,
		&auction.SellerID,
		&auction.AuctionType,
		&auction.Title,
		&auction.Description,
		&auction.StartingPrice,
		&auction.CurrentPrice,
		&auction.ReservePrice,
		&auction.BuyNowPrice,
		&auction.Currency,
		&auction.MinIncrement,
		&auction.PriceDecrement,
		&auction.DecrementIntervalSecs,
		&auction.Status,
		&auction.StartsAt,
		&auction.EndsAt,
		&auction.ExtensionSeconds,
		&auction.WinningBidID,
		&auction.WinnerID,
		&auction.Metadata,
		&auction.CreatedAt,
		&auction.UpdatedAt,
		&auction.BidCount,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAuctionNotFound
		}
		return nil, err
	}

	return &auction, nil
}

// SearchAuctions searches for auctions.
func (r *Repository) SearchAuctions(ctx context.Context, params SearchAuctionsParams) (*AuctionListResult, error) {
	baseQuery := `
		SELECT
			a.id, a.listing_id, a.seller_id, a.auction_type, a.title, a.description,
			a.starting_price, a.current_price, a.reserve_price, a.buy_now_price, a.price_currency,
			a.min_increment, a.price_decrement, a.decrement_interval_seconds,
			a.status, a.starts_at, a.ends_at, a.extension_seconds,
			a.winning_bid_id, a.winner_id, a.metadata, a.created_at, a.updated_at,
			(SELECT COUNT(*) FROM bids WHERE auction_id = a.id) as bid_count
		FROM auctions a
		WHERE 1=1
	`

	countQuery := `SELECT COUNT(*) FROM auctions a WHERE 1=1`
	args := []any{}
	argNum := 1

	if params.SellerID != nil {
		baseQuery += ` AND a.seller_id = $` + string(rune('0'+argNum))
		countQuery += ` AND a.seller_id = $` + string(rune('0'+argNum))
		args = append(args, *params.SellerID)
		argNum++
	}

	if params.AuctionType != nil {
		baseQuery += ` AND a.auction_type = $` + string(rune('0'+argNum))
		countQuery += ` AND a.auction_type = $` + string(rune('0'+argNum))
		args = append(args, *params.AuctionType)
		argNum++
	}

	if params.Status != nil {
		baseQuery += ` AND a.status = $` + string(rune('0'+argNum))
		countQuery += ` AND a.status = $` + string(rune('0'+argNum))
		args = append(args, *params.Status)
		argNum++
	}

	if params.Query != "" {
		baseQuery += ` AND (a.title ILIKE $` + string(rune('0'+argNum)) + ` OR a.description ILIKE $` + string(rune('0'+argNum)) + `)`
		countQuery += ` AND (a.title ILIKE $` + string(rune('0'+argNum)) + ` OR a.description ILIKE $` + string(rune('0'+argNum)) + `)`
		args = append(args, "%"+params.Query+"%")
		argNum++
	}

	// Get total count
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Add pagination
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	baseQuery += ` ORDER BY a.created_at DESC`
	baseQuery += ` LIMIT $` + string(rune('0'+argNum)) + ` OFFSET $` + string(rune('0'+argNum+1))
	args = append(args, params.Limit, params.Offset)

	rows, err := r.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var auctions []*Auction
	for rows.Next() {
		var auction Auction
		if err := rows.Scan(
			&auction.ID,
			&auction.ListingID,
			&auction.SellerID,
			&auction.AuctionType,
			&auction.Title,
			&auction.Description,
			&auction.StartingPrice,
			&auction.CurrentPrice,
			&auction.ReservePrice,
			&auction.BuyNowPrice,
			&auction.Currency,
			&auction.MinIncrement,
			&auction.PriceDecrement,
			&auction.DecrementIntervalSecs,
			&auction.Status,
			&auction.StartsAt,
			&auction.EndsAt,
			&auction.ExtensionSeconds,
			&auction.WinningBidID,
			&auction.WinnerID,
			&auction.Metadata,
			&auction.CreatedAt,
			&auction.UpdatedAt,
			&auction.BidCount,
		); err != nil {
			return nil, err
		}
		auctions = append(auctions, &auction)
	}

	return &AuctionListResult{
		Auctions: auctions,
		Total:    total,
		Limit:    params.Limit,
		Offset:   params.Offset,
	}, nil
}

// CreateBid creates a new bid.
func (r *Repository) CreateBid(ctx context.Context, bid *Bid) error {
	query := `
		INSERT INTO bids (id, auction_id, bidder_id, amount, currency, is_sealed, status, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		bid.ID,
		bid.AuctionID,
		bid.BidderID,
		bid.Amount,
		bid.Currency,
		bid.IsSealed,
		bid.Status,
		bid.Metadata,
		bid.CreatedAt,
	)

	return err
}

// GetBidsByAuctionID retrieves all bids for an auction.
func (r *Repository) GetBidsByAuctionID(ctx context.Context, auctionID uuid.UUID) ([]*Bid, error) {
	query := `
		SELECT id, auction_id, bidder_id, amount, currency, is_sealed, status, metadata, created_at
		FROM bids
		WHERE auction_id = $1
		ORDER BY amount DESC, created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, auctionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []*Bid
	for rows.Next() {
		var bid Bid
		if err := rows.Scan(
			&bid.ID,
			&bid.AuctionID,
			&bid.BidderID,
			&bid.Amount,
			&bid.Currency,
			&bid.IsSealed,
			&bid.Status,
			&bid.Metadata,
			&bid.CreatedAt,
		); err != nil {
			return nil, err
		}
		bids = append(bids, &bid)
	}

	return bids, nil
}

// GetHighestBid retrieves the highest bid for an auction.
func (r *Repository) GetHighestBid(ctx context.Context, auctionID uuid.UUID) (*Bid, error) {
	query := `
		SELECT id, auction_id, bidder_id, amount, currency, is_sealed, status, metadata, created_at
		FROM bids
		WHERE auction_id = $1 AND status IN ('active', 'winning')
		ORDER BY amount DESC
		LIMIT 1
	`

	var bid Bid
	err := r.pool.QueryRow(ctx, query, auctionID).Scan(
		&bid.ID,
		&bid.AuctionID,
		&bid.BidderID,
		&bid.Amount,
		&bid.Currency,
		&bid.IsSealed,
		&bid.Status,
		&bid.Metadata,
		&bid.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No bids yet
		}
		return nil, err
	}

	return &bid, nil
}

// UpdateBidStatus updates the status of a bid.
func (r *Repository) UpdateBidStatus(ctx context.Context, bidID uuid.UUID, status BidStatus) error {
	query := `UPDATE bids SET status = $2 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, bidID, status)
	return err
}

// UpdateAuctionPrice updates the current price of an auction.
func (r *Repository) UpdateAuctionPrice(ctx context.Context, auctionID uuid.UUID, price float64) error {
	query := `UPDATE auctions SET current_price = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, auctionID, price)
	return err
}

// UpdateAuctionStatus updates the status of an auction.
func (r *Repository) UpdateAuctionStatus(ctx context.Context, auctionID uuid.UUID, status AuctionStatus) error {
	query := `UPDATE auctions SET status = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, auctionID, status)
	return err
}

// SetAuctionWinner sets the winning bid and winner of an auction.
func (r *Repository) SetAuctionWinner(ctx context.Context, auctionID, winningBidID, winnerID uuid.UUID) error {
	query := `
		UPDATE auctions
		SET winning_bid_id = $2, winner_id = $3, status = 'ended', updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, auctionID, winningBidID, winnerID)
	return err
}

// ExtendAuction extends the end time of an auction (anti-sniping).
func (r *Repository) ExtendAuction(ctx context.Context, auctionID uuid.UUID, newEndTime time.Time) error {
	query := `UPDATE auctions SET ends_at = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, auctionID, newEndTime)
	return err
}

// MarkPreviousBidsOutbid marks all previous bids as outbid.
func (r *Repository) MarkPreviousBidsOutbid(ctx context.Context, auctionID uuid.UUID, exceptBidID uuid.UUID) error {
	query := `UPDATE bids SET status = 'outbid' WHERE auction_id = $1 AND id != $2 AND status = 'active'`
	_, err := r.pool.Exec(ctx, query, auctionID, exceptBidID)
	return err
}
