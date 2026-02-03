package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/jwt"
	clerkuser "github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/digi604/swarmmarket/backend/internal/agent"
	"github.com/digi604/swarmmarket/backend/internal/common"
	"github.com/digi604/swarmmarket/backend/internal/user"
	"github.com/google/uuid"
)

// ContextKey is a type for context keys.
type ContextKey string

const (
	// AgentContextKey is the context key for the authenticated agent.
	AgentContextKey ContextKey = "agent"
)

// Auth creates an authentication middleware.
func Auth(agentService *agent.Service, apiKeyHeader string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get(apiKeyHeader)
			if apiKey == "" {
				// Also check Authorization header with Bearer scheme
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					apiKey = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if apiKey == "" {
				// Use generic error to prevent information leakage
				common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("unauthorized"))
				return
			}

			ag, err := agentService.ValidateAPIKey(r.Context(), apiKey)
			if err != nil {
				// Use generic error for all auth failures to prevent enumeration attacks
				common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("unauthorized"))
				return
			}

			// Add agent to context
			ctx := context.WithValue(r.Context(), AgentContextKey, ag)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAgent retrieves the authenticated agent from the request context.
func GetAgent(ctx context.Context) *agent.Agent {
	ag, ok := ctx.Value(AgentContextKey).(*agent.Agent)
	if !ok {
		return nil
	}
	return ag
}

// OptionalAuth creates an optional authentication middleware.
// It attempts to authenticate but doesn't fail if no credentials are provided.
func OptionalAuth(agentService *agent.Service, apiKeyHeader string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get(apiKeyHeader)
			if apiKey == "" {
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					apiKey = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if apiKey != "" {
				ag, err := agentService.ValidateAPIKey(r.Context(), apiKey)
				if err == nil {
					ctx := context.WithValue(r.Context(), AgentContextKey, ag)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CombinedAuth creates middleware that accepts EITHER:
// 1. Agent API key (X-API-Key or Bearer token starting with "sm_")
// 2. Clerk JWT + X-Act-As-Agent header (human acting as their owned agent)
func CombinedAuth(agentService *agent.Service, userService *user.Service, apiKeyHeader string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try API key first (for agents)
			apiKey := r.Header.Get(apiKeyHeader)
			if apiKey == "" {
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					token := strings.TrimPrefix(authHeader, "Bearer ")
					// Only treat as API key if it starts with "sm_"
					if strings.HasPrefix(token, "sm_") {
						apiKey = token
					}
				}
			}

			if apiKey != "" && strings.HasPrefix(apiKey, "sm_") {
				ag, err := agentService.ValidateAPIKey(r.Context(), apiKey)
				if err == nil {
					ctx := context.WithValue(r.Context(), AgentContextKey, ag)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			// Try Clerk JWT + X-Act-As-Agent (for humans acting as their owned agent)
			authHeader := r.Header.Get("Authorization")
			actAsAgentID := r.Header.Get("X-Act-As-Agent")

			if strings.HasPrefix(authHeader, "Bearer ") && actAsAgentID != "" {
				token := strings.TrimPrefix(authHeader, "Bearer ")

				// Don't process API keys as JWTs
				if strings.HasPrefix(token, "sm_") {
					common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("unauthorized"))
					return
				}

				// Verify Clerk JWT
				claims, err := jwt.Verify(r.Context(), &jwt.VerifyParams{Token: token})
				if err != nil {
					common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("invalid token"))
					return
				}

				// Get user from Clerk
				clerkUser, err := clerkuser.Get(r.Context(), claims.Subject)
				if err != nil {
					common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("user not found"))
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

				// Sync user to database
				usr, err := userService.GetOrCreateUser(r.Context(), &user.ClerkUser{
					ID:           clerkUser.ID,
					EmailAddress: email,
					FirstName:    stringVal(clerkUser.FirstName),
					LastName:     stringVal(clerkUser.LastName),
					ImageURL:     stringVal(clerkUser.ImageURL),
				})
				if err != nil {
					common.WriteError(w, http.StatusInternalServerError, common.ErrInternalServer("failed to sync user"))
					return
				}

				// Parse agent ID
				agentID, err := uuid.Parse(actAsAgentID)
				if err != nil {
					common.WriteError(w, http.StatusBadRequest, common.ErrBadRequest("invalid agent ID"))
					return
				}

				// Get the agent and verify ownership
				ag, err := agentService.GetByID(r.Context(), agentID)
				if err != nil {
					common.WriteError(w, http.StatusNotFound, common.ErrNotFound("agent not found"))
					return
				}

				// Verify the user owns this agent
				if ag.OwnerUserID == nil || *ag.OwnerUserID != usr.ID {
					common.WriteError(w, http.StatusForbidden, common.ErrForbidden("you do not own this agent"))
					return
				}

				// Set agent in context (handlers will see the agent, not the user)
				ctx := context.WithValue(r.Context(), AgentContextKey, ag)
				ctx = context.WithValue(ctx, UserContextKey, usr) // Also set user for reference
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			common.WriteError(w, http.StatusUnauthorized, common.ErrUnauthorized("unauthorized"))
		})
	}
}

// stringVal safely dereferences a string pointer.
func stringVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
