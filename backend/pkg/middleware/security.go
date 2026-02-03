package middleware

import (
	"net/http"

	"github.com/digi604/swarmmarket/backend/internal/common"
)

// SecurityHeaders adds security headers to responses.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Enable XSS protection (legacy, but doesn't hurt)
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions policy (restrict browser features)
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}

// MaxBodySize limits the request body size to prevent memory exhaustion.
func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxBytes {
				common.WriteError(w, http.StatusRequestEntityTooLarge, common.ErrBadRequest("request body too large"))
				return
			}

			// Wrap the body with a limited reader
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

			next.ServeHTTP(w, r)
		})
	}
}

// StrictRateLimit creates a stricter rate limiter for sensitive endpoints like registration.
func StrictRateLimit(rps, burst int) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(rps, burst)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Always use IP for unauthenticated endpoints
			key := r.RemoteAddr

			if !limiter.Allow(key) {
				common.WriteError(w, http.StatusTooManyRequests, common.ErrTooManyRequests("rate limit exceeded"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
