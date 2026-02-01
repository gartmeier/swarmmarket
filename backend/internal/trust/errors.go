package trust

import "errors"

var (
	// Verification errors
	ErrVerificationNotFound   = errors.New("verification not found")
	ErrAlreadyVerified        = errors.New("already verified")
	ErrChallengeNotFound      = errors.New("challenge not found")
	ErrChallengeExpired       = errors.New("challenge expired")
	ErrChallengeNotPending    = errors.New("challenge is not pending")
	ErrMaxAttemptsExceeded    = errors.New("maximum verification attempts exceeded")
	ErrInvalidChallengeID     = errors.New("invalid challenge id")
	ErrUnauthorized           = errors.New("unauthorized")

	// Twitter verification errors
	ErrTwitterNotConfigured   = errors.New("twitter verification not configured")
	ErrInvalidTweetURL        = errors.New("invalid tweet URL")
	ErrTweetNotFound          = errors.New("tweet not found")
	ErrTweetTextMismatch      = errors.New("tweet does not contain verification text")
	ErrTwitterAPIError        = errors.New("twitter API error")
)
