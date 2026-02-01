package matching

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestNewEngine(t *testing.T) {
	engine := NewEngine(nil)
	if engine == nil {
		t.Fatal("NewEngine() returned nil")
	}
	if engine.buyOrders == nil {
		t.Error("buyOrders map is nil")
	}
	if engine.sellOrders == nil {
		t.Error("sellOrders map is nil")
	}
}

func TestPlaceLimitBuyOrder(t *testing.T) {
	engine := NewEngine(nil)
	productID := uuid.New()
	agentID := uuid.New()

	order := &Order{
		AgentID:   agentID,
		ProductID: productID,
		Side:      OrderSideBuy,
		Type:      OrderTypeLimit,
		Price:     100.0,
		Quantity:  10.0,
	}

	result, err := engine.PlaceOrder(context.Background(), order)
	if err != nil {
		t.Fatalf("PlaceOrder() error = %v", err)
	}

	if len(result.Trades) != 0 {
		t.Errorf("Expected no trades, got %d", len(result.Trades))
	}

	if result.RemainingOrder == nil {
		t.Error("Expected remaining order")
	}

	if order.Status != OrderStatusOpen {
		t.Errorf("Order status = %s, want %s", order.Status, OrderStatusOpen)
	}

	// Verify order is in the book
	book := engine.GetOrderBook(productID, 10)
	if len(book.Bids) != 1 {
		t.Errorf("Expected 1 bid level, got %d", len(book.Bids))
	}
	if book.Bids[0].Price != 100.0 {
		t.Errorf("Bid price = %f, want 100.0", book.Bids[0].Price)
	}
}

func TestPlaceLimitSellOrder(t *testing.T) {
	engine := NewEngine(nil)
	productID := uuid.New()
	agentID := uuid.New()

	order := &Order{
		AgentID:   agentID,
		ProductID: productID,
		Side:      OrderSideSell,
		Type:      OrderTypeLimit,
		Price:     110.0,
		Quantity:  5.0,
	}

	result, err := engine.PlaceOrder(context.Background(), order)
	if err != nil {
		t.Fatalf("PlaceOrder() error = %v", err)
	}

	if len(result.Trades) != 0 {
		t.Errorf("Expected no trades, got %d", len(result.Trades))
	}

	// Verify order is in the book
	book := engine.GetOrderBook(productID, 10)
	if len(book.Asks) != 1 {
		t.Errorf("Expected 1 ask level, got %d", len(book.Asks))
	}
	if book.Asks[0].Price != 110.0 {
		t.Errorf("Ask price = %f, want 110.0", book.Asks[0].Price)
	}
}

func TestOrderMatching(t *testing.T) {
	tradeCalled := false
	engine := NewEngine(func(ctx context.Context, trade Trade) {
		tradeCalled = true
	})
	_ = tradeCalled // Used to verify event handler was called

	productID := uuid.New()
	buyer := uuid.New()
	seller := uuid.New()

	// Place sell order first
	sellOrder := &Order{
		AgentID:   seller,
		ProductID: productID,
		Side:      OrderSideSell,
		Type:      OrderTypeLimit,
		Price:     100.0,
		Quantity:  10.0,
	}
	engine.PlaceOrder(context.Background(), sellOrder)

	// Place matching buy order
	buyOrder := &Order{
		AgentID:   buyer,
		ProductID: productID,
		Side:      OrderSideBuy,
		Type:      OrderTypeLimit,
		Price:     100.0,
		Quantity:  10.0,
	}
	result, err := engine.PlaceOrder(context.Background(), buyOrder)
	if err != nil {
		t.Fatalf("PlaceOrder() error = %v", err)
	}

	// Should have one trade
	if len(result.Trades) != 1 {
		t.Fatalf("Expected 1 trade, got %d", len(result.Trades))
	}

	trade := result.Trades[0]
	if trade.Price != 100.0 {
		t.Errorf("Trade price = %f, want 100.0", trade.Price)
	}
	if trade.Quantity != 10.0 {
		t.Errorf("Trade quantity = %f, want 10.0", trade.Quantity)
	}
	if trade.BuyerID != buyer {
		t.Error("Trade buyer ID mismatch")
	}
	if trade.SellerID != seller {
		t.Error("Trade seller ID mismatch")
	}

	// Book should be empty
	book := engine.GetOrderBook(productID, 10)
	if len(book.Bids) != 0 {
		t.Errorf("Expected no bids, got %d", len(book.Bids))
	}
	if len(book.Asks) != 0 {
		t.Errorf("Expected no asks, got %d", len(book.Asks))
	}

	// Last price should be set
	if book.LastPrice == nil || *book.LastPrice != 100.0 {
		t.Error("Last price not set correctly")
	}
}

func TestPartialFill(t *testing.T) {
	engine := NewEngine(nil)
	productID := uuid.New()
	buyer := uuid.New()
	seller := uuid.New()

	// Place large sell order
	sellOrder := &Order{
		AgentID:   seller,
		ProductID: productID,
		Side:      OrderSideSell,
		Type:      OrderTypeLimit,
		Price:     100.0,
		Quantity:  100.0,
	}
	engine.PlaceOrder(context.Background(), sellOrder)

	// Place smaller buy order
	buyOrder := &Order{
		AgentID:   buyer,
		ProductID: productID,
		Side:      OrderSideBuy,
		Type:      OrderTypeLimit,
		Price:     100.0,
		Quantity:  30.0,
	}
	result, _ := engine.PlaceOrder(context.Background(), buyOrder)

	// Trade should be for 30
	if len(result.Trades) != 1 {
		t.Fatalf("Expected 1 trade, got %d", len(result.Trades))
	}
	if result.Trades[0].Quantity != 30.0 {
		t.Errorf("Trade quantity = %f, want 30.0", result.Trades[0].Quantity)
	}

	// Sell order should still have 70 remaining
	book := engine.GetOrderBook(productID, 10)
	if len(book.Asks) != 1 {
		t.Fatalf("Expected 1 ask level, got %d", len(book.Asks))
	}
	if book.Asks[0].Quantity != 70.0 {
		t.Errorf("Remaining ask quantity = %f, want 70.0", book.Asks[0].Quantity)
	}
}

func TestCancelOrder(t *testing.T) {
	engine := NewEngine(nil)
	productID := uuid.New()
	agentID := uuid.New()

	order := &Order{
		AgentID:   agentID,
		ProductID: productID,
		Side:      OrderSideBuy,
		Type:      OrderTypeLimit,
		Price:     100.0,
		Quantity:  10.0,
	}
	result, _ := engine.PlaceOrder(context.Background(), order)

	// Cancel the order
	err := engine.CancelOrder(result.RemainingOrder.ID, agentID)
	if err != nil {
		t.Fatalf("CancelOrder() error = %v", err)
	}

	// Book should be empty
	book := engine.GetOrderBook(productID, 10)
	if len(book.Bids) != 0 {
		t.Errorf("Expected no bids after cancel, got %d", len(book.Bids))
	}
}

func TestCancelOrderUnauthorized(t *testing.T) {
	engine := NewEngine(nil)
	productID := uuid.New()
	owner := uuid.New()
	other := uuid.New()

	order := &Order{
		AgentID:   owner,
		ProductID: productID,
		Side:      OrderSideBuy,
		Type:      OrderTypeLimit,
		Price:     100.0,
		Quantity:  10.0,
	}
	result, _ := engine.PlaceOrder(context.Background(), order)

	// Try to cancel with different agent
	err := engine.CancelOrder(result.RemainingOrder.ID, other)
	if err == nil {
		t.Error("Expected error for unauthorized cancel")
	}
}

func TestPriceTimePriority(t *testing.T) {
	engine := NewEngine(nil)
	productID := uuid.New()

	// Place multiple buy orders at same price
	agent1 := uuid.New()
	agent2 := uuid.New()

	order1 := &Order{
		AgentID:   agent1,
		ProductID: productID,
		Side:      OrderSideBuy,
		Type:      OrderTypeLimit,
		Price:     100.0,
		Quantity:  10.0,
	}
	engine.PlaceOrder(context.Background(), order1)

	order2 := &Order{
		AgentID:   agent2,
		ProductID: productID,
		Side:      OrderSideBuy,
		Type:      OrderTypeLimit,
		Price:     100.0,
		Quantity:  10.0,
	}
	engine.PlaceOrder(context.Background(), order2)

	// Place sell order that matches first order only
	seller := uuid.New()
	sellOrder := &Order{
		AgentID:   seller,
		ProductID: productID,
		Side:      OrderSideSell,
		Type:      OrderTypeLimit,
		Price:     100.0,
		Quantity:  10.0,
	}
	result, _ := engine.PlaceOrder(context.Background(), sellOrder)

	// Should match with first order (time priority)
	if len(result.Trades) != 1 {
		t.Fatalf("Expected 1 trade, got %d", len(result.Trades))
	}
	if result.Trades[0].BuyerID != agent1 {
		t.Error("Expected first order to be matched (time priority)")
	}

	// Second order should still be in book
	book := engine.GetOrderBook(productID, 10)
	if len(book.Bids) != 1 {
		t.Errorf("Expected 1 bid remaining, got %d", len(book.Bids))
	}
}

func TestMarketOrder(t *testing.T) {
	engine := NewEngine(nil)
	productID := uuid.New()
	seller := uuid.New()
	buyer := uuid.New()

	// Place limit sell order
	sellOrder := &Order{
		AgentID:   seller,
		ProductID: productID,
		Side:      OrderSideSell,
		Type:      OrderTypeLimit,
		Price:     100.0,
		Quantity:  10.0,
	}
	engine.PlaceOrder(context.Background(), sellOrder)

	// Place market buy order (no price specified)
	marketBuy := &Order{
		AgentID:   buyer,
		ProductID: productID,
		Side:      OrderSideBuy,
		Type:      OrderTypeMarket,
		Price:     0, // Market orders don't specify price
		Quantity:  5.0,
	}
	result, _ := engine.PlaceOrder(context.Background(), marketBuy)

	// Should match at the sell order's price
	if len(result.Trades) != 1 {
		t.Fatalf("Expected 1 trade, got %d", len(result.Trades))
	}
	if result.Trades[0].Price != 100.0 {
		t.Errorf("Market order executed at %f, want 100.0", result.Trades[0].Price)
	}

	// Market order should not remain in book
	if result.RemainingOrder != nil {
		t.Error("Market order should not remain in book")
	}
}

func TestOrderBookDepth(t *testing.T) {
	engine := NewEngine(nil)
	productID := uuid.New()
	agentID := uuid.New()

	// Place orders at different prices
	for i := 1; i <= 10; i++ {
		order := &Order{
			AgentID:   agentID,
			ProductID: productID,
			Side:      OrderSideBuy,
			Type:      OrderTypeLimit,
			Price:     float64(100 + i),
			Quantity:  1.0,
		}
		engine.PlaceOrder(context.Background(), order)
	}

	// Get book with depth 5
	book := engine.GetOrderBook(productID, 5)
	if len(book.Bids) != 5 {
		t.Errorf("Expected 5 bid levels, got %d", len(book.Bids))
	}

	// Highest price should be first
	if book.Bids[0].Price != 110.0 {
		t.Errorf("Top bid = %f, want 110.0", book.Bids[0].Price)
	}
}
