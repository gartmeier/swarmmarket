package trust

import "math"

// Trust score constants
const (
	BaseTrustScore          = 0.5
	MaxTrustScore           = 1.0
	OwnershipTrustScore     = 1.0  // Claimed agents get instant max trust
	TwitterTrustBonus       = 0.15
	MaxTransactionTrustBonus = 0.25
	MaxRatingTrustBonus     = 0.10
	TransactionDecayRate    = 0.05
	MinRatingsForBonus      = 5
)

// TransactionTrustBonus calculates trust from successful transactions using exponential decay.
// Formula: bonus = maxBonus * (1 - e^(-decayRate * transactions))
// This gives diminishing returns: early transactions worth more, later ones worth less.
//
// Example values:
//   - 1 transaction:   +0.01
//   - 10 transactions: +0.10
//   - 50 transactions: +0.23
//   - 100 transactions: +0.25 (approaching max)
func TransactionTrustBonus(successfulTransactions int) float64 {
	if successfulTransactions <= 0 {
		return 0
	}

	bonus := MaxTransactionTrustBonus * (1 - math.Exp(-TransactionDecayRate*float64(successfulTransactions)))
	return roundTo4Decimals(bonus)
}

// RatingTrustBonus calculates trust from average rating.
// Only applies after minimum number of ratings (5).
// Rating of 3.0 or below gives 0 bonus.
// Rating of 5.0 gives maximum bonus (0.10).
func RatingTrustBonus(averageRating float64, ratingCount int) float64 {
	if ratingCount < MinRatingsForBonus || averageRating <= 3.0 {
		return 0
	}

	// Scale from 3.0 (0 bonus) to 5.0 (max bonus)
	// (rating - 3.0) / 2.0 gives 0 to 1 range
	bonus := MaxRatingTrustBonus * ((averageRating - 3.0) / 2.0)
	return math.Min(MaxRatingTrustBonus, roundTo4Decimals(bonus))
}

// CalculateTotalTrustScore computes the total trust score from components.
// If isOwnerClaimed is true, returns instant max trust (1.0).
func CalculateTotalTrustScore(isOwnerClaimed bool, verificationBonus, transactionBonus, ratingBonus float64) float64 {
	if isOwnerClaimed {
		return OwnershipTrustScore
	}

	total := BaseTrustScore + verificationBonus + transactionBonus + ratingBonus
	return math.Min(MaxTrustScore, roundTo4Decimals(total))
}

// roundTo4Decimals rounds a float64 to 4 decimal places
func roundTo4Decimals(value float64) float64 {
	return math.Round(value*10000) / 10000
}
