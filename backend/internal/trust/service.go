package trust

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service handles trust score and verification logic.
type Service struct {
	repo    *Repository
	twitter *TwitterVerifier
}

// NewService creates a new trust service.
func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// SetTwitterVerifier sets the Twitter verifier.
func (s *Service) SetTwitterVerifier(twitter *TwitterVerifier) {
	s.twitter = twitter
}

// --- Twitter Verification ---

// InitiateTwitterVerification starts the Twitter verification process.
// Returns a challenge text that the agent must tweet.
func (s *Service) InitiateTwitterVerification(ctx context.Context, agentID uuid.UUID, agentName string) (*InitiateTwitterVerificationResponse, error) {
	// Check if already verified
	existing, err := s.repo.GetVerificationByAgentAndType(ctx, agentID, VerificationTwitter)
	if err == nil && existing.Status == StatusVerified {
		return nil, ErrAlreadyVerified
	}

	// Generate unique challenge text (viral marketing tweet)
	agentIDPrefix := agentID.String()[:8]
	challengeText := fmt.Sprintf(
		"I just registered my AI agent on @SwarmMarket - the autonomous agent marketplace where AIs trade goods, services, and data.\n\nVerifying: %s #SwarmMarket #AIAgents\n\nhttps://swarmmarket.ai",
		agentIDPrefix,
	)

	// Create or update verification record
	var verification *AgentVerification
	if existing != nil {
		verification = existing
		verification.Status = StatusPending
		if err := s.repo.UpdateVerification(ctx, verification); err != nil {
			return nil, err
		}
	} else {
		verification = &AgentVerification{
			AgentID:          agentID,
			VerificationType: VerificationTwitter,
			Status:           StatusPending,
		}
		if err := s.repo.CreateVerification(ctx, verification); err != nil {
			return nil, err
		}
	}

	// Create challenge
	challenge := &VerificationChallenge{
		AgentID:        agentID,
		VerificationID: &verification.ID,
		ChallengeType:  "twitter_post",
		ChallengeText:  agentIDPrefix, // We only need to verify the unique prefix is in the tweet
		MaxAttempts:    5,
		Status:         "pending",
		ExpiresAt:      time.Now().UTC().Add(24 * time.Hour),
	}
	if err := s.repo.CreateChallenge(ctx, challenge); err != nil {
		return nil, err
	}

	return &InitiateTwitterVerificationResponse{
		ChallengeID:   challenge.ID.String(),
		ChallengeText: challengeText,
		Instructions:  "Post a tweet containing the exact text above, then provide the tweet URL to confirm verification.",
		ExpiresAt:     challenge.ExpiresAt,
	}, nil
}

// ConfirmTwitterVerification confirms Twitter verification with a tweet URL.
func (s *Service) ConfirmTwitterVerification(ctx context.Context, agentID uuid.UUID, req *ConfirmTwitterVerificationRequest) (*ConfirmVerificationResponse, error) {
	if s.twitter == nil {
		return nil, ErrTwitterNotConfigured
	}

	challengeID, err := uuid.Parse(req.ChallengeID)
	if err != nil {
		return nil, ErrInvalidChallengeID
	}

	// Get challenge
	challenge, err := s.repo.GetChallengeByID(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	// Validate ownership
	if challenge.AgentID != agentID {
		return nil, ErrUnauthorized
	}

	// Check status
	if challenge.Status != "pending" {
		return nil, ErrChallengeNotPending
	}

	// Check expiry
	if time.Now().After(challenge.ExpiresAt) {
		s.repo.UpdateChallengeStatus(ctx, challengeID, "expired")
		return nil, ErrChallengeExpired
	}

	// Check attempts
	if challenge.Attempts >= challenge.MaxAttempts {
		s.repo.UpdateChallengeStatus(ctx, challengeID, "failed")
		return nil, ErrMaxAttemptsExceeded
	}

	// Store the tweet URL
	s.repo.UpdateChallengeTweetURL(ctx, challengeID, req.TweetURL)

	// Verify tweet via Twitter API
	twitterHandle, twitterUserID, err := s.twitter.VerifyTweet(ctx, req.TweetURL, challenge.ChallengeText)
	if err != nil {
		s.repo.IncrementChallengeAttempts(ctx, challengeID)
		return &ConfirmVerificationResponse{
			Verified: false,
			Message:  fmt.Sprintf("Verification failed: %v", err),
		}, nil
	}

	// Mark challenge as verified
	s.repo.UpdateChallengeStatus(ctx, challengeID, "verified")

	// Update verification record
	verification, err := s.repo.GetVerificationByAgentAndType(ctx, agentID, VerificationTwitter)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	verification.Status = StatusVerified
	verification.TwitterHandle = twitterHandle
	verification.TwitterUserID = twitterUserID
	verification.VerificationTweetID = extractTweetID(req.TweetURL)
	verification.TrustBonus = TwitterTrustBonus
	verification.VerifiedAt = &now
	if err := s.repo.UpdateVerification(ctx, verification); err != nil {
		return nil, err
	}

	// Recalculate trust score
	newScore, err := s.RecalculateTrustScore(ctx, agentID)
	if err != nil {
		return nil, err
	}

	return &ConfirmVerificationResponse{
		Verified:      true,
		TrustBonus:    TwitterTrustBonus,
		NewTrustScore: newScore,
		Message:       fmt.Sprintf("Twitter account @%s verified successfully!", twitterHandle),
	}, nil
}

// --- Trust Score Calculation ---

// RecalculateTrustScore recalculates and updates an agent's trust score.
func (s *Service) RecalculateTrustScore(ctx context.Context, agentID uuid.UUID) (float64, error) {
	// Get current agent data
	currentScore, isOwnerClaimed, successfulTrades, avgRating, err := s.repo.GetAgentTrustData(ctx, agentID)
	if err != nil {
		return 0, err
	}

	// If claimed by human owner, trust is always 1.0
	if isOwnerClaimed {
		if currentScore != OwnershipTrustScore {
			s.repo.UpdateAgentTrustComponents(ctx, agentID, OwnershipTrustScore, 0, 0, 0)
		}
		return OwnershipTrustScore, nil
	}

	// Get verifications
	verifications, _ := s.repo.GetVerificationsByAgent(ctx, agentID)

	// Calculate verification bonus
	verificationBonus := 0.0
	for _, v := range verifications {
		if v.Status == StatusVerified {
			verificationBonus += v.TrustBonus
		}
	}

	// Calculate transaction bonus (diminishing returns)
	transactionBonus := TransactionTrustBonus(successfulTrades)

	// Calculate rating bonus
	ratingCount, _ := s.repo.GetRatingCount(ctx, agentID)
	ratingBonus := RatingTrustBonus(avgRating, ratingCount)

	// Calculate total
	newScore := CalculateTotalTrustScore(false, verificationBonus, transactionBonus, ratingBonus)

	// Update agent
	if err := s.repo.UpdateAgentTrustComponents(ctx, agentID, newScore, verificationBonus, transactionBonus, ratingBonus); err != nil {
		return 0, err
	}

	// Record history if score changed
	if newScore != currentScore {
		change := newScore - currentScore
		reason := ReasonTransactionCompleted // Default, caller should use specific methods
		s.repo.RecordTrustChange(ctx, &TrustScoreHistory{
			AgentID:       agentID,
			PreviousScore: currentScore,
			NewScore:      newScore,
			ChangeReason:  reason,
			ChangeAmount:  change,
		})
	}

	return newScore, nil
}

// OnTransactionCompleted should be called when a transaction completes successfully.
func (s *Service) OnTransactionCompleted(ctx context.Context, agentID uuid.UUID, transactionID uuid.UUID) error {
	currentScore, _, _, _, err := s.repo.GetAgentTrustData(ctx, agentID)
	if err != nil {
		return err
	}

	newScore, err := s.RecalculateTrustScore(ctx, agentID)
	if err != nil {
		return err
	}

	change := newScore - currentScore
	if change > 0 {
		s.repo.RecordTrustChange(ctx, &TrustScoreHistory{
			AgentID:       agentID,
			PreviousScore: currentScore,
			NewScore:      newScore,
			ChangeReason:  ReasonTransactionCompleted,
			ChangeAmount:  change,
			Metadata:      map[string]any{"transaction_id": transactionID.String()},
		})
	}

	return nil
}

// OnRatingReceived should be called when an agent receives a rating.
func (s *Service) OnRatingReceived(ctx context.Context, agentID uuid.UUID, ratingScore int, transactionID uuid.UUID) error {
	currentScore, _, _, _, err := s.repo.GetAgentTrustData(ctx, agentID)
	if err != nil {
		return err
	}

	newScore, err := s.RecalculateTrustScore(ctx, agentID)
	if err != nil {
		return err
	}

	change := newScore - currentScore
	s.repo.RecordTrustChange(ctx, &TrustScoreHistory{
		AgentID:       agentID,
		PreviousScore: currentScore,
		NewScore:      newScore,
		ChangeReason:  ReasonRatingReceived,
		ChangeAmount:  change,
		Metadata:      map[string]any{"rating_score": ratingScore, "transaction_id": transactionID.String()},
	})

	return nil
}

// OnOwnershipClaimed should be called when an agent is claimed by a human owner.
func (s *Service) OnOwnershipClaimed(ctx context.Context, agentID uuid.UUID, userID uuid.UUID) error {
	currentScore, _, _, _, err := s.repo.GetAgentTrustData(ctx, agentID)
	if err != nil {
		return err
	}

	// Ownership gives instant max trust
	newScore := OwnershipTrustScore
	change := newScore - currentScore

	s.repo.RecordTrustChange(ctx, &TrustScoreHistory{
		AgentID:       agentID,
		PreviousScore: currentScore,
		NewScore:      newScore,
		ChangeReason:  ReasonOwnershipClaimed,
		ChangeAmount:  change,
		Metadata:      map[string]any{"user_id": userID.String()},
	})

	return nil
}

// --- Trust Breakdown ---

// GetTrustBreakdown returns the detailed trust score breakdown for an agent.
func (s *Service) GetTrustBreakdown(ctx context.Context, agentID uuid.UUID) (*TrustBreakdown, error) {
	currentScore, isOwnerClaimed, successfulTrades, avgRating, err := s.repo.GetAgentTrustData(ctx, agentID)
	if err != nil {
		return nil, err
	}

	verifications, _ := s.repo.GetVerificationsByAgent(ctx, agentID)
	ratingCount, _ := s.repo.GetRatingCount(ctx, agentID)

	// Build verification summaries
	verificationSummaries := make([]VerificationSummary, 0, len(verifications))
	verificationBonus := 0.0
	for _, v := range verifications {
		summary := VerificationSummary{
			Type:       v.VerificationType,
			Status:     v.Status,
			TrustBonus: v.TrustBonus,
			VerifiedAt: v.VerifiedAt,
		}
		if v.TwitterHandle != "" {
			summary.Handle = v.TwitterHandle
		}
		verificationSummaries = append(verificationSummaries, summary)
		if v.Status == StatusVerified {
			verificationBonus += v.TrustBonus
		}
	}

	transactionBonus := TransactionTrustBonus(successfulTrades)
	ratingBonus := RatingTrustBonus(avgRating, ratingCount)

	// Get total transactions (not just successful)
	// For now, we'll use successful trades as a proxy
	totalTransactions := successfulTrades

	return &TrustBreakdown{
		AgentID:           agentID,
		TotalScore:        currentScore,
		BaseScore:         BaseTrustScore,
		VerificationBonus: verificationBonus,
		TransactionBonus:  transactionBonus,
		RatingBonus:       ratingBonus,
		IsOwnerClaimed:    isOwnerClaimed,
		Verifications:     verificationSummaries,
		TransactionCount:  totalTransactions,
		SuccessfulTrades:  successfulTrades,
		AverageRating:     avgRating,
		RatingCount:       ratingCount,
	}, nil
}

// GetTrustHistory returns the trust score change history for an agent.
func (s *Service) GetTrustHistory(ctx context.Context, agentID uuid.UUID, limit int) ([]*TrustScoreHistory, error) {
	return s.repo.GetTrustHistory(ctx, agentID, limit)
}

// GetVerifications returns all verifications for an agent.
func (s *Service) GetVerifications(ctx context.Context, agentID uuid.UUID) ([]*AgentVerification, error) {
	return s.repo.GetVerificationsByAgent(ctx, agentID)
}
