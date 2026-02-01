package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/jwt"
	clerkuser "github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/digi604/swarmmarket/backend/internal/common"
	"github.com/digi604/swarmmarket/backend/internal/user"
)

const UserContextKey ContextKey = "user"

// ClerkAuth creates middleware that validates Clerk JWTs and syncs users.
func ClerkAuth(userService *user.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("missing authorization token"))
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Verify JWT with Clerk
			claims, err := jwt.Verify(r.Context(), &jwt.VerifyParams{Token: token})
			if err != nil {
				common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("invalid token"))
				return
			}

			// Get user from Clerk
			clerkUser, err := clerkuser.Get(r.Context(), claims.Subject)
			if err != nil {
				common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("user not found in Clerk"))
				return
			}

			// Get primary email
			var email string
			for _, emailAddr := range clerkUser.EmailAddresses {
				if emailAddr.ID == *clerkUser.PrimaryEmailAddressID {
					email = emailAddr.EmailAddress
					break
				}
			}
			if email == "" && len(clerkUser.EmailAddresses) > 0 {
				email = clerkUser.EmailAddresses[0].EmailAddress
			}

			// Upsert user in our database
			usr, err := userService.GetOrCreateUser(r.Context(), &user.ClerkUser{
				ID:           clerkUser.ID,
				EmailAddress: email,
				FirstName:    stringValue(clerkUser.FirstName),
				LastName:     stringValue(clerkUser.LastName),
				ImageURL:     stringValue(clerkUser.ImageURL),
			})
			if err != nil {
				common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to sync user"))
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, usr)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalClerkAuth creates optional Clerk authentication middleware.
// It attempts to authenticate but doesn't fail if no credentials are provided.
func OptionalClerkAuth(userService *user.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				// No auth header, continue without user
				next.ServeHTTP(w, r)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Verify JWT with Clerk
			claims, err := jwt.Verify(r.Context(), &jwt.VerifyParams{Token: token})
			if err != nil {
				// Invalid token, continue without user
				next.ServeHTTP(w, r)
				return
			}

			// Get user from Clerk
			clerkUser, err := clerkuser.Get(r.Context(), claims.Subject)
			if err != nil {
				// User not found, continue without user
				next.ServeHTTP(w, r)
				return
			}

			// Get primary email
			var email string
			for _, emailAddr := range clerkUser.EmailAddresses {
				if emailAddr.ID == *clerkUser.PrimaryEmailAddressID {
					email = emailAddr.EmailAddress
					break
				}
			}
			if email == "" && len(clerkUser.EmailAddresses) > 0 {
				email = clerkUser.EmailAddresses[0].EmailAddress
			}

			// Upsert user in our database
			usr, err := userService.GetOrCreateUser(r.Context(), &user.ClerkUser{
				ID:           clerkUser.ID,
				EmailAddress: email,
				FirstName:    stringValue(clerkUser.FirstName),
				LastName:     stringValue(clerkUser.LastName),
				ImageURL:     stringValue(clerkUser.ImageURL),
			})
			if err != nil {
				// Failed to sync user, continue without user
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, usr)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUser retrieves the authenticated user from the request context.
func GetUser(ctx context.Context) *user.User {
	u, ok := ctx.Value(UserContextKey).(*user.User)
	if !ok {
		return nil
	}
	return u
}

// stringValue safely dereferences a string pointer.
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
