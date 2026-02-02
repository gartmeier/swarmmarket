package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/digi604/swarmmarket/backend/internal/auction"
	"github.com/digi604/swarmmarket/backend/internal/common"
	"github.com/digi604/swarmmarket/backend/pkg/middleware"
)

// AuctionHandler handles auction HTTP requests.
type AuctionHandler struct {
	service *auction.Service
}

// NewAuctionHandler creates a new auction handler.
func NewAuctionHandler(service *auction.Service) *AuctionHandler {
	return &AuctionHandler{service: service}
}

// CreateAuction handles POST /auctions - create a new auction.
func (h *AuctionHandler) CreateAuction(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	var req auction.CreateAuctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	if req.Title == "" {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("title is required"))
		return
	}

	if req.AuctionType == "" {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("auction_type is required"))
		return
	}

	auc, err := h.service.CreateAuction(r.Context(), agent.ID, &req)
	if err != nil {
		switch err {
		case auction.ErrInvalidAuctionType:
			common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid auction type - must be english, dutch, sealed, or continuous"))
		default:
			common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest(err.Error()))
		}
		return
	}

	common.WriteJSON(w, http.StatusCreated, auc)
}

// GetAuction handles GET /auctions/{id} - get auction details.
func (h *AuctionHandler) GetAuction(w http.ResponseWriter, r *http.Request) {
	idOrSlug := chi.URLParam(r, "id")

	var auc *auction.Auction
	var err error

	// Try parsing as UUID first
	if id, parseErr := uuid.Parse(idOrSlug); parseErr == nil {
		auc, err = h.service.GetAuction(r.Context(), id)
	} else {
		// Fall back to slug lookup
		auc, err = h.service.GetAuctionBySlug(r.Context(), idOrSlug)
	}

	if err != nil {
		if err == auction.ErrAuctionNotFound {
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("auction not found"))
			return
		}
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to get auction"))
		return
	}

	common.WriteJSON(w, http.StatusOK, auc)
}

// SearchAuctions handles GET /auctions - search auctions.
func (h *AuctionHandler) SearchAuctions(w http.ResponseWriter, r *http.Request) {
	params := auction.SearchAuctionsParams{
		Query:  r.URL.Query().Get("q"),
		Limit:  parseIntParam(r, "limit", 20),
		Offset: parseIntParam(r, "offset", 0),
	}

	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		auctionType := auction.AuctionType(typeStr)
		params.AuctionType = &auctionType
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := auction.AuctionStatus(statusStr)
		params.Status = &status
	}

	if sellerStr := r.URL.Query().Get("seller_id"); sellerStr != "" {
		if sellerID, err := uuid.Parse(sellerStr); err == nil {
			params.SellerID = &sellerID
		}
	}

	result, err := h.service.SearchAuctions(r.Context(), params)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to search auctions"))
		return
	}

	common.WriteJSON(w, http.StatusOK, result)
}

// PlaceBid handles POST /auctions/{id}/bid - place a bid.
func (h *AuctionHandler) PlaceBid(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	idStr := chi.URLParam(r, "id")
	auctionID, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid auction id"))
		return
	}

	var req auction.PlaceBidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	if req.Amount <= 0 {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("amount must be positive"))
		return
	}

	bid, err := h.service.PlaceBid(r.Context(), auctionID, agent.ID, &req)
	if err != nil {
		switch err {
		case auction.ErrAuctionNotFound:
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("auction not found"))
		case auction.ErrAuctionNotActive:
			common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("auction is not active"))
		case auction.ErrAuctionEnded:
			common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("auction has ended"))
		case auction.ErrBidTooLow:
			common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("bid amount is too low"))
		case auction.ErrCannotBidOnOwnAuction:
			common.WriteError(w, http.StatusForbidden, common.ErrForbidden("cannot bid on your own auction"))
		default:
			common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to place bid"))
		}
		return
	}

	common.WriteJSON(w, http.StatusCreated, bid)
}

// GetBids handles GET /auctions/{id}/bids - get bids for an auction.
func (h *AuctionHandler) GetBids(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	auctionID, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid auction id"))
		return
	}

	// Get optional requester ID for sealed auctions
	var requesterID *uuid.UUID
	if agent := middleware.GetAgent(r.Context()); agent != nil {
		requesterID = &agent.ID
	}

	bids, err := h.service.GetBids(r.Context(), auctionID, requesterID)
	if err != nil {
		if err == auction.ErrAuctionNotFound {
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("auction not found"))
			return
		}
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to get bids"))
		return
	}

	common.WriteJSON(w, http.StatusOK, map[string]any{
		"bids":  bids,
		"total": len(bids),
	})
}

// EndAuction handles POST /auctions/{id}/end - end an auction.
func (h *AuctionHandler) EndAuction(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	idStr := chi.URLParam(r, "id")
	auctionID, err := uuid.Parse(idStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid auction id"))
		return
	}

	auc, err := h.service.EndAuction(r.Context(), auctionID, agent.ID)
	if err != nil {
		switch err {
		case auction.ErrAuctionNotFound:
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound("auction not found"))
		case auction.ErrAuctionNotActive:
			common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("auction is not active"))
		case auction.ErrNotAuthorized:
			common.WriteError(w, http.StatusForbidden, common.ErrForbidden("only the seller can end the auction"))
		default:
			common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to end auction"))
		}
		return
	}

	common.WriteJSON(w, http.StatusOK, auc)
}
