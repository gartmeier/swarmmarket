package payment

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/refund"
	"github.com/stripe/stripe-go/v76/transfer"
)

var (
	ErrPaymentFailed     = errors.New("payment failed")
	ErrRefundFailed      = errors.New("refund failed")
	ErrTransferFailed    = errors.New("transfer failed")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrInvalidCurrency   = errors.New("invalid currency")
)

// Config holds Stripe configuration.
type Config struct {
	SecretKey      string
	WebhookSecret  string
	PlatformFeePercent float64 // e.g., 0.025 for 2.5%
}

// Service handles Stripe payments for escrow.
type Service struct {
	config Config
}

// NewService creates a new payment service.
func NewService(cfg Config) *Service {
	stripe.Key = cfg.SecretKey
	return &Service{config: cfg}
}

// CreateEscrowPayment creates a payment intent for escrow.
// The funds are held until released to the seller.
func (s *Service) CreateEscrowPayment(ctx context.Context, req *CreatePaymentRequest) (*PaymentResult, error) {
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Convert to cents
	amountCents := int64(req.Amount * 100)

	// Calculate platform fee
	platformFee := int64(float64(amountCents) * s.config.PlatformFeePercent)

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountCents),
		Currency: stripe.String(normalizeCurrency(req.Currency)),
		Metadata: map[string]string{
			"transaction_id": req.TransactionID.String(),
			"buyer_id":       req.BuyerID.String(),
			"seller_id":      req.SellerID.String(),
		},
		CaptureMethod: stripe.String("manual"), // Hold funds, capture later
	}

	// If seller has a connected Stripe account, set up for direct transfer
	if req.SellerStripeAccountID != "" {
		params.TransferData = &stripe.PaymentIntentTransferDataParams{
			Destination: stripe.String(req.SellerStripeAccountID),
		}
		params.ApplicationFeeAmount = stripe.Int64(platformFee)
	}

	intent, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPaymentFailed, err)
	}

	return &PaymentResult{
		PaymentIntentID: intent.ID,
		ClientSecret:    intent.ClientSecret,
		Status:          string(intent.Status),
		Amount:          req.Amount,
		Currency:        req.Currency,
	}, nil
}

// CapturePayment captures a held payment (releases from escrow to seller).
func (s *Service) CapturePayment(ctx context.Context, paymentIntentID string) error {
	_, err := paymentintent.Capture(paymentIntentID, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPaymentFailed, err)
	}
	return nil
}

// RefundPayment refunds a payment (for disputes or cancellations).
func (s *Service) RefundPayment(ctx context.Context, paymentIntentID string, amount *float64) error {
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentID),
	}

	// Partial refund if amount specified
	if amount != nil {
		params.Amount = stripe.Int64(int64(*amount * 100))
	}

	_, err := refund.New(params)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRefundFailed, err)
	}
	return nil
}

// TransferToSeller transfers funds directly to seller's connected account.
// Use this when not using PaymentIntent with transfer_data.
func (s *Service) TransferToSeller(ctx context.Context, req *TransferRequest) (*TransferResult, error) {
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	amountCents := int64(req.Amount * 100)

	params := &stripe.TransferParams{
		Amount:      stripe.Int64(amountCents),
		Currency:    stripe.String(normalizeCurrency(req.Currency)),
		Destination: stripe.String(req.SellerStripeAccountID),
		Metadata: map[string]string{
			"transaction_id": req.TransactionID.String(),
		},
	}

	if req.SourceTransactionID != "" {
		params.SourceTransaction = stripe.String(req.SourceTransactionID)
	}

	xfer, err := transfer.New(params)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransferFailed, err)
	}

	return &TransferResult{
		TransferID: xfer.ID,
		Amount:     req.Amount,
		Currency:   req.Currency,
		Status:     "completed",
	}, nil
}

// GetPaymentIntent retrieves a payment intent.
func (s *Service) GetPaymentIntent(ctx context.Context, paymentIntentID string) (*PaymentStatus, error) {
	intent, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		return nil, err
	}

	return &PaymentStatus{
		PaymentIntentID: intent.ID,
		Status:          string(intent.Status),
		Amount:          float64(intent.Amount) / 100,
		Currency:        string(intent.Currency),
		CapturedAmount:  float64(intent.AmountReceived) / 100,
	}, nil
}

// CreatePaymentRequest is the request to create an escrow payment.
type CreatePaymentRequest struct {
	TransactionID         uuid.UUID
	BuyerID               uuid.UUID
	SellerID              uuid.UUID
	Amount                float64
	Currency              string
	SellerStripeAccountID string // Optional: seller's connected Stripe account
}

// PaymentResult is the result of creating a payment.
type PaymentResult struct {
	PaymentIntentID string  `json:"payment_intent_id"`
	ClientSecret    string  `json:"client_secret"`
	Status          string  `json:"status"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
}

// PaymentStatus is the status of a payment.
type PaymentStatus struct {
	PaymentIntentID string  `json:"payment_intent_id"`
	Status          string  `json:"status"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	CapturedAmount  float64 `json:"captured_amount"`
}

// TransferRequest is the request to transfer funds to seller.
type TransferRequest struct {
	TransactionID         uuid.UUID
	SellerStripeAccountID string
	Amount                float64
	Currency              string
	SourceTransactionID   string // Original charge ID for connected transfers
}

// TransferResult is the result of a transfer.
type TransferResult struct {
	TransferID string  `json:"transfer_id"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	Status     string  `json:"status"`
}

func normalizeCurrency(currency string) string {
	if currency == "" {
		return "usd"
	}
	// Stripe requires lowercase
	switch currency {
	case "USD", "usd":
		return "usd"
	case "EUR", "eur":
		return "eur"
	case "GBP", "gbp":
		return "gbp"
	default:
		return "usd"
	}
}
