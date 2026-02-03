package trust

import "math"

// Trust score constants (0-100% scale, stored as 0.0-1.0)
const (
	BaseTrustScore           = 0.0  // New agents start at 0%
	MaxTrustScore            = 1.0  // 100% max
	HumanLinkBonus           = 0.10 // +10% for linking to human owner
	TwitterTrustBonus        = 0.15 // +15% for Twitter verification
	MaxTransactionTrustBonus = 0.75 // +75% max from successful transactions
	TransactionDecayRate     = 0.03 // Slower decay for more gradual growth
)

// TransactionTrustBonus calculates trust from successful transactions using exponential decay.
// Formula: bonus = maxBonus * (1 - e^(-decayRate * transactions))
// This gives diminishing returns: early transactions worth more, later ones worth less.
//
// Example values (with 75% max, 0.03 decay):
//   - 1 transaction:   +2%
//   - 10 transactions: +22%
//   - 25 transactions: +42%
//   - 50 transactions: +55%
//   - 100 transactions: +70%
func TransactionTrustBonus(successfulTransactions int) float64 {
	if successfulTransactions <= 0 {
		return 0
	}

	bonus := MaxTransactionTrustBonus * (1 - math.Exp(-TransactionDecayRate*float64(successfulTransactions)))
	return roundTo4Decimals(bonus)
}

// CalculateTotalTrustScore computes the total trust score from components.
// Trust = Base (0%) + Human Link (+10%) + Verifications (+15%) + Transactions (up to +75%)
// Max possible: 100%
func CalculateTotalTrustScore(isOwnerClaimed bool, verificationBonus, transactionBonus float64) float64 {
	total := BaseTrustScore + transactionBonus + verificationBonus

	// Human-linked agents get +10% bonus
	if isOwnerClaimed {
		total += HumanLinkBonus
	}

	return math.Min(MaxTrustScore, roundTo4Decimals(total))
}

// roundTo4Decimals rounds a float64 to 4 decimal places
func roundTo4Decimals(value float64) float64 {
	return math.Round(value*10000) / 10000
}
