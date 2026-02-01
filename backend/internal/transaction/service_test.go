package transaction

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// mockRepository implements a mock repository for testing.
type mockRepository struct {
	transactions map[uuid.UUID]*Transaction
	escrows      map[uuid.UUID]*EscrowAccount
	ratings      map[uuid.UUID][]*Rating
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		transactions: make(map[uuid.UUID]*Transaction),
		escrows:      make(map[uuid.UUID]*EscrowAccount),
		ratings:      make(map[uuid.UUID][]*Rating),
	}
}

// mockPublisher implements EventPublisher for testing.
type mockPublisher struct {
	events []publishedEvent
}

type publishedEvent struct {
	eventType string
	payload   map[string]any
}

func (m *mockPublisher) Publish(ctx context.Context, eventType string, payload map[string]any) error {
	m.events = append(m.events, publishedEvent{eventType, payload})
	return nil
}

// mockPaymentService implements PaymentService for testing.
type mockPaymentService struct {
	paymentIntents map[string]bool
	captured       []string
	refunded       []string
}

func newMockPaymentService() *mockPaymentService {
	return &mockPaymentService{
		paymentIntents: make(map[string]bool),
	}
}

func (m *mockPaymentService) CreateEscrowPayment(ctx context.Context, transactionID, buyerID, sellerID string, amount float64, currency string) (string, string, error) {
	piID := "pi_test_" + transactionID[:8]
	m.paymentIntents[piID] = true
	return piID, piID + "_secret", nil
}

func (m *mockPaymentService) CapturePayment(ctx context.Context, paymentIntentID string) error {
	m.captured = append(m.captured, paymentIntentID)
	return nil
}

func (m *mockPaymentService) RefundPayment(ctx context.Context, paymentIntentID string) error {
	m.refunded = append(m.refunded, paymentIntentID)
	return nil
}

func TestTransactionStatus(t *testing.T) {
	tests := []struct {
		status   TransactionStatus
		expected string
	}{
		{StatusPending, "pending"},
		{StatusEscrowFunded, "escrow_funded"},
		{StatusDelivered, "delivered"},
		{StatusCompleted, "completed"},
		{StatusDisputed, "disputed"},
		{StatusRefunded, "refunded"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.expected {
			t.Errorf("expected status %s, got %s", tt.expected, tt.status)
		}
	}
}

func TestEscrowStatus(t *testing.T) {
	tests := []struct {
		status   EscrowStatus
		expected string
	}{
		{EscrowPending, "pending"},
		{EscrowFunded, "funded"},
		{EscrowReleased, "released"},
		{EscrowRefunded, "refunded"},
		{EscrowDisputed, "disputed"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.expected {
			t.Errorf("expected status %s, got %s", tt.expected, tt.status)
		}
	}
}

func TestEscrowFundingResult(t *testing.T) {
	result := &EscrowFundingResult{
		TransactionID:   uuid.New(),
		PaymentIntentID: "pi_test123",
		ClientSecret:    "pi_test123_secret",
		Amount:          100.00,
		Currency:        "USD",
	}

	if result.PaymentIntentID != "pi_test123" {
		t.Errorf("expected payment intent id pi_test123, got %s", result.PaymentIntentID)
	}

	if result.Amount != 100.00 {
		t.Errorf("expected amount 100.00, got %f", result.Amount)
	}
}

func TestCreateTransactionRequest(t *testing.T) {
	buyerID := uuid.New()
	sellerID := uuid.New()
	requestID := uuid.New()
	offerID := uuid.New()

	req := &CreateTransactionRequest{
		BuyerID:   buyerID,
		SellerID:  sellerID,
		RequestID: &requestID,
		OfferID:   &offerID,
		Amount:    50.00,
		Currency:  "USD",
	}

	if req.BuyerID != buyerID {
		t.Errorf("expected buyer id %s, got %s", buyerID, req.BuyerID)
	}

	if req.Amount != 50.00 {
		t.Errorf("expected amount 50.00, got %f", req.Amount)
	}
}

func TestSubmitRatingRequest(t *testing.T) {
	// Valid rating
	req := &SubmitRatingRequest{
		Score:   5,
		Comment: "Excellent service!",
	}

	if req.Score < 1 || req.Score > 5 {
		t.Error("rating score should be between 1 and 5")
	}

	// Test boundary values
	validScores := []int{1, 2, 3, 4, 5}
	for _, score := range validScores {
		if score < 1 || score > 5 {
			t.Errorf("score %d should be valid", score)
		}
	}
}

func TestDisputeRequest(t *testing.T) {
	req := &DisputeRequest{
		Reason:      "Item not as described",
		Description: "The data provided was incomplete and missing key fields.",
	}

	if req.Reason == "" {
		t.Error("dispute reason should not be empty")
	}
}

func TestListTransactionsParams(t *testing.T) {
	agentID := uuid.New()
	status := StatusPending

	params := ListTransactionsParams{
		AgentID: &agentID,
		Status:  &status,
		Role:    "buyer",
		Limit:   20,
		Offset:  0,
	}

	if *params.AgentID != agentID {
		t.Errorf("expected agent id %s, got %s", agentID, *params.AgentID)
	}

	if *params.Status != StatusPending {
		t.Errorf("expected status pending, got %s", *params.Status)
	}

	if params.Role != "buyer" {
		t.Errorf("expected role buyer, got %s", params.Role)
	}
}

func TestTransactionListResult(t *testing.T) {
	result := &TransactionListResult{
		Items:  []*Transaction{},
		Total:  0,
		Limit:  20,
		Offset: 0,
	}

	if result.Limit != 20 {
		t.Errorf("expected limit 20, got %d", result.Limit)
	}

	if len(result.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(result.Items))
	}
}

func TestTransaction(t *testing.T) {
	buyerID := uuid.New()
	sellerID := uuid.New()

	tx := &Transaction{
		ID:       uuid.New(),
		BuyerID:  buyerID,
		SellerID: sellerID,
		Amount:   100.00,
		Currency: "USD",
		Status:   StatusPending,
	}

	if tx.BuyerID != buyerID {
		t.Errorf("expected buyer id %s, got %s", buyerID, tx.BuyerID)
	}

	if tx.Status != StatusPending {
		t.Errorf("expected status pending, got %s", tx.Status)
	}
}

func TestEscrowAccount(t *testing.T) {
	txID := uuid.New()

	escrow := &EscrowAccount{
		ID:            uuid.New(),
		TransactionID: txID,
		Amount:        100.00,
		Currency:      "USD",
		Status:        EscrowPending,
	}

	if escrow.TransactionID != txID {
		t.Errorf("expected transaction id %s, got %s", txID, escrow.TransactionID)
	}

	if escrow.Status != EscrowPending {
		t.Errorf("expected status pending, got %s", escrow.Status)
	}
}

func TestRating(t *testing.T) {
	txID := uuid.New()
	raterID := uuid.New()
	ratedID := uuid.New()

	rating := &Rating{
		ID:            uuid.New(),
		TransactionID: txID,
		RaterID:       raterID,
		RatedAgentID:  ratedID,
		Score:         5,
		Comment:       "Great transaction!",
	}

	if rating.Score != 5 {
		t.Errorf("expected score 5, got %d", rating.Score)
	}

	if rating.RaterID != raterID {
		t.Errorf("expected rater id %s, got %s", raterID, rating.RaterID)
	}
}

func TestServiceErrors(t *testing.T) {
	// Test error messages
	if ErrInvalidStatus.Error() != "invalid transaction status for this operation" {
		t.Errorf("unexpected error message: %s", ErrInvalidStatus.Error())
	}

	if ErrNotAuthorized.Error() != "not authorized to perform this action" {
		t.Errorf("unexpected error message: %s", ErrNotAuthorized.Error())
	}

	if ErrInvalidRating.Error() != "rating score must be between 1 and 5" {
		t.Errorf("unexpected error message: %s", ErrInvalidRating.Error())
	}

	if ErrCannotRateYourself.Error() != "cannot rate yourself" {
		t.Errorf("unexpected error message: %s", ErrCannotRateYourself.Error())
	}

	if ErrTransactionNotReady.Error() != "transaction is not ready for this operation" {
		t.Errorf("unexpected error message: %s", ErrTransactionNotReady.Error())
	}
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
	if service.payment != nil {
		t.Error("payment should be nil initially")
	}
}

func TestSetPaymentService(t *testing.T) {
	service := NewService(nil, nil)
	paymentService := newMockPaymentService()

	service.SetPaymentService(paymentService)

	if service.payment == nil {
		t.Error("payment service should be set")
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

func TestTransactionAllFields(t *testing.T) {
	now := time.Now()
	requestID := uuid.New()
	offerID := uuid.New()
	listingID := uuid.New()
	auctionID := uuid.New()
	completedAt := now.Add(24 * time.Hour)
	deliveryConfirmedAt := now.Add(12 * time.Hour)

	tx := &Transaction{
		ID:                  uuid.New(),
		BuyerID:             uuid.New(),
		SellerID:            uuid.New(),
		ListingID:           &listingID,
		RequestID:           &requestID,
		OfferID:             &offerID,
		AuctionID:           &auctionID,
		Amount:              500.0,
		Currency:            "USD",
		PlatformFee:         25.0,
		Status:              StatusCompleted,
		DeliveryConfirmedAt: &deliveryConfirmedAt,
		CompletedAt:         &completedAt,
		Metadata:            map[string]any{"notes": "test transaction"},
		CreatedAt:           now,
		UpdatedAt:           now,
		BuyerName:           "Test Buyer",
		SellerName:          "Test Seller",
	}

	if tx.Amount != 500.0 {
		t.Errorf("expected amount 500.0, got %f", tx.Amount)
	}
	if tx.Status != StatusCompleted {
		t.Error("status not set correctly")
	}
	if tx.PlatformFee != 25.0 {
		t.Errorf("expected platform fee 25.0, got %f", tx.PlatformFee)
	}
	if tx.BuyerName != "Test Buyer" {
		t.Error("buyer name not set correctly")
	}
	if tx.DeliveryConfirmedAt == nil {
		t.Error("delivery confirmed at should be set")
	}
}

func TestEscrowAccountAllFields(t *testing.T) {
	now := time.Now()
	releasedAt := now.Add(48 * time.Hour)

	escrow := &EscrowAccount{
		ID:                    uuid.New(),
		TransactionID:         uuid.New(),
		Amount:                250.0,
		Currency:              "EUR",
		Status:                EscrowFunded,
		StripePaymentIntentID: "pi_test123",
		FundedAt:              &now,
		ReleasedAt:            &releasedAt,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if escrow.Amount != 250.0 {
		t.Errorf("expected amount 250.0, got %f", escrow.Amount)
	}
	if escrow.Status != EscrowFunded {
		t.Error("status not set correctly")
	}
	if escrow.StripePaymentIntentID != "pi_test123" {
		t.Error("payment intent ID not set correctly")
	}
	if escrow.FundedAt == nil {
		t.Error("funded at should be set")
	}
}

func TestRatingAllFields(t *testing.T) {
	now := time.Now()
	rating := &Rating{
		ID:            uuid.New(),
		TransactionID: uuid.New(),
		RaterID:       uuid.New(),
		RatedAgentID:  uuid.New(),
		Score:         5,
		Comment:       "Excellent service, highly recommended!",
		CreatedAt:     now,
	}

	if rating.Score != 5 {
		t.Errorf("expected score 5, got %d", rating.Score)
	}
	if rating.Comment != "Excellent service, highly recommended!" {
		t.Error("comment not set correctly")
	}
}

func TestRatingScoreValidation(t *testing.T) {
	validScores := []int{1, 2, 3, 4, 5}
	for _, score := range validScores {
		if score < 1 || score > 5 {
			t.Errorf("score %d should be valid", score)
		}
	}

	invalidScores := []int{0, -1, 6, 10, 100}
	for _, score := range invalidScores {
		if score >= 1 && score <= 5 {
			t.Errorf("score %d should be invalid", score)
		}
	}
}

func TestListTransactionsParamsDefaults(t *testing.T) {
	params := ListTransactionsParams{}

	if params.Limit != 0 {
		t.Errorf("expected default limit 0, got %d", params.Limit)
	}
	if params.Offset != 0 {
		t.Errorf("expected default offset 0, got %d", params.Offset)
	}
}

func TestListTransactionsParamsWithFilters(t *testing.T) {
	agentID := uuid.New()
	status := StatusDelivered

	params := ListTransactionsParams{
		AgentID: &agentID,
		Status:  &status,
		Role:    "seller",
		Limit:   50,
		Offset:  100,
	}

	if *params.AgentID != agentID {
		t.Error("agent ID not set correctly")
	}
	if *params.Status != StatusDelivered {
		t.Error("status not set correctly")
	}
	if params.Role != "seller" {
		t.Errorf("expected role 'seller', got %s", params.Role)
	}
	if params.Limit != 50 {
		t.Errorf("expected limit 50, got %d", params.Limit)
	}
	if params.Offset != 100 {
		t.Errorf("expected offset 100, got %d", params.Offset)
	}
}

func TestTransactionListResultEmpty(t *testing.T) {
	result := &TransactionListResult{
		Items:  []*Transaction{},
		Total:  0,
		Limit:  20,
		Offset: 0,
	}

	if len(result.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(result.Items))
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
}

func TestTransactionListResultWithItems(t *testing.T) {
	items := []*Transaction{
		{ID: uuid.New(), Amount: 100.0},
		{ID: uuid.New(), Amount: 200.0},
		{ID: uuid.New(), Amount: 300.0},
	}

	result := &TransactionListResult{
		Items:  items,
		Total:  100,
		Limit:  20,
		Offset: 40,
	}

	if len(result.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(result.Items))
	}
	if result.Total != 100 {
		t.Errorf("expected total 100, got %d", result.Total)
	}
	if result.Offset != 40 {
		t.Errorf("expected offset 40, got %d", result.Offset)
	}
}

func TestCreateTransactionRequestValidation(t *testing.T) {
	requestID := uuid.New()
	offerID := uuid.New()

	req := &CreateTransactionRequest{
		BuyerID:   uuid.New(),
		SellerID:  uuid.New(),
		RequestID: &requestID,
		OfferID:   &offerID,
		Amount:    100.0,
		Currency:  "USD",
	}

	// Validate required fields
	if req.BuyerID == uuid.Nil {
		t.Error("buyer ID should not be nil")
	}
	if req.SellerID == uuid.Nil {
		t.Error("seller ID should not be nil")
	}
	if req.Amount <= 0 {
		t.Error("amount should be positive")
	}
}

func TestSubmitRatingRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     SubmitRatingRequest
		wantErr bool
	}{
		{
			name:    "valid rating",
			req:     SubmitRatingRequest{Score: 5, Comment: "Great!"},
			wantErr: false,
		},
		{
			name:    "score too low",
			req:     SubmitRatingRequest{Score: 0},
			wantErr: true,
		},
		{
			name:    "score too high",
			req:     SubmitRatingRequest{Score: 6},
			wantErr: true,
		},
		{
			name:    "min valid score",
			req:     SubmitRatingRequest{Score: 1},
			wantErr: false,
		},
		{
			name:    "max valid score",
			req:     SubmitRatingRequest{Score: 5},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasErr := tt.req.Score < 1 || tt.req.Score > 5
			if hasErr != tt.wantErr {
				t.Errorf("validation error = %v, wantErr %v", hasErr, tt.wantErr)
			}
		})
	}
}

func TestDisputeRequestValidation(t *testing.T) {
	req := &DisputeRequest{
		Reason:      "Item not received",
		Description: "I never received the data package that was promised.",
	}

	if req.Reason == "" {
		t.Error("reason should not be empty")
	}
	if len(req.Description) == 0 {
		t.Error("description should be provided")
	}
}

func TestDisputeRequestEmptyReason(t *testing.T) {
	req := &DisputeRequest{
		Reason:      "",
		Description: "Some description",
	}

	if req.Reason != "" {
		t.Error("expected empty reason")
	}
}

func TestEscrowFundingResultAllFields(t *testing.T) {
	result := &EscrowFundingResult{
		TransactionID:   uuid.New(),
		PaymentIntentID: "pi_test_abc123",
		ClientSecret:    "pi_test_abc123_secret_xyz",
		Amount:          150.0,
		Currency:        "USD",
	}

	if result.PaymentIntentID != "pi_test_abc123" {
		t.Error("payment intent ID not set correctly")
	}
	if result.ClientSecret != "pi_test_abc123_secret_xyz" {
		t.Error("client secret not set correctly")
	}
	if result.Amount != 150.0 {
		t.Errorf("expected amount 150.0, got %f", result.Amount)
	}
}
