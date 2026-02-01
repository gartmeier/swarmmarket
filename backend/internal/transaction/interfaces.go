package transaction

import (
	"context"

	"github.com/google/uuid"
)

// RepositoryInterface defines the contract for transaction data persistence.
// This interface enables mock implementations for testing.
type RepositoryInterface interface {
	// Transaction Operations
	CreateTransaction(ctx context.Context, req *CreateTransactionRequest) (*Transaction, error)
	GetTransactionByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	ListTransactions(ctx context.Context, params ListTransactionsParams) (*TransactionListResult, error)
	UpdateTransactionStatus(ctx context.Context, id uuid.UUID, status TransactionStatus) error
	ConfirmDelivery(ctx context.Context, id uuid.UUID) error
	CompleteTransaction(ctx context.Context, id uuid.UUID) error

	// Escrow Operations
	CreateEscrowAccount(ctx context.Context, transactionID uuid.UUID, amount float64, currency string) (*EscrowAccount, error)
	GetEscrowByTransactionID(ctx context.Context, transactionID uuid.UUID) (*EscrowAccount, error)
	UpdateEscrowStatus(ctx context.Context, id uuid.UUID, status EscrowStatus) error
	UpdateEscrowPaymentIntent(ctx context.Context, id uuid.UUID, paymentIntentID string) error

	// Rating Operations
	CreateRating(ctx context.Context, rating *Rating) error
	GetRatingsByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*Rating, error)
	HasRated(ctx context.Context, transactionID, raterID uuid.UUID) (bool, error)

	// Agent Stats
	UpdateAgentStats(ctx context.Context, agentID uuid.UUID, successful bool) error
	RecalculateAgentRating(ctx context.Context, agentID uuid.UUID) error
}

// Verify that Repository implements RepositoryInterface
var _ RepositoryInterface = (*Repository)(nil)
