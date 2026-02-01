package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/digi604/swarmmarket/backend/internal/common"
	"github.com/digi604/swarmmarket/backend/internal/matching"
	"github.com/digi604/swarmmarket/backend/pkg/middleware"
)

// OrderBookHandler handles order book HTTP requests.
type OrderBookHandler struct {
	engine *matching.Engine
}

// NewOrderBookHandler creates a new order book handler.
func NewOrderBookHandler(engine *matching.Engine) *OrderBookHandler {
	return &OrderBookHandler{engine: engine}
}

// PlaceOrderRequest is the request body for placing an order.
type PlaceOrderRequest struct {
	ProductID string  `json:"product_id"`
	Side      string  `json:"side"`     // "buy" or "sell"
	Type      string  `json:"type"`     // "limit" or "market"
	Price     float64 `json:"price"`    // Required for limit orders
	Quantity  float64 `json:"quantity"`
}

// PlaceOrder handles POST /orderbook/orders - place a new order.
func (h *OrderBookHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	var req PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid request body"))
		return
	}

	// Validate
	if req.ProductID == "" {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("product_id is required"))
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid product_id"))
		return
	}

	if req.Side != "buy" && req.Side != "sell" {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("side must be 'buy' or 'sell'"))
		return
	}

	if req.Type != "limit" && req.Type != "market" {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("type must be 'limit' or 'market'"))
		return
	}

	if req.Type == "limit" && req.Price <= 0 {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("price is required for limit orders"))
		return
	}

	if req.Quantity <= 0 {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("quantity must be positive"))
		return
	}

	order := &matching.Order{
		AgentID:   agent.ID,
		ProductID: productID,
		Side:      matching.OrderSide(req.Side),
		Type:      matching.OrderType(req.Type),
		Price:     req.Price,
		Quantity:  req.Quantity,
	}

	result, err := h.engine.PlaceOrder(r.Context(), order)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to place order"))
		return
	}

	common.WriteJSON(w, http.StatusCreated, map[string]any{
		"order":  order,
		"trades": result.Trades,
	})
}

// GetOrderBook handles GET /orderbook/{productId} - get order book.
func (h *OrderBookHandler) GetOrderBook(w http.ResponseWriter, r *http.Request) {
	productIDStr := chi.URLParam(r, "productId")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid product_id"))
		return
	}

	depth := parseIntParam(r, "depth", 10)
	if depth > 50 {
		depth = 50
	}

	book := h.engine.GetOrderBook(productID, depth)
	common.WriteJSON(w, http.StatusOK, book)
}

// CancelOrder handles DELETE /orderbook/orders/{orderId} - cancel an order.
func (h *OrderBookHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	agent := middleware.GetAgent(r.Context())
	if agent == nil {
		common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("not authenticated"))
		return
	}

	orderIDStr := chi.URLParam(r, "orderId")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid order_id"))
		return
	}

	if err := h.engine.CancelOrder(orderID, agent.ID); err != nil {
		if err.Error() == "not authorized to cancel this order" {
			common.WriteError(w, http.StatusForbidden, common.ErrForbidden(err.Error()))
			return
		}
		if err.Error() == "order not found" {
			common.WriteError(w, http.StatusNotFound, common.ErrNotFound(err.Error()))
			return
		}
		common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to cancel order"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
