package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/thecodearcher/limen"
)

type contextKey string
type requestMetadataKey struct{}
type RequestMetadata struct {
	UserID string
}
type ExecGroupChecker interface {
	HasExecGroup(ctx context.Context, userID string) (bool, error)
}

const sessionKey contextKey = "session"

// Requires user to be signed in to access endpoints under this
func RequireAuth(auth *limen.Limen) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := auth.GetSession(r)
			if err != nil || session == nil {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
				return
			}

			if session.User != nil {
				if userID, ok := session.User.ID.(string); ok {
					if metadata := RequestMetadataFromContext(r.Context()); metadata != nil {
						metadata.UserID = userID
					}
				}
			}

			ctx := context.WithValue(r.Context(), sessionKey, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Requires user to have these roles to access endpoints under this
// eg. r.Use(auth.RequireRole("admin")) or r.Use(auth.RequireRole("admin", "member"))
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := SessionFromContext(r.Context())
			if session == nil {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
				return
			}
			role, _ := session.User.Raw()["role"].(string)
			if !allowed[role] {
				writeError(w, http.StatusForbidden, "FORBIDDEN", "Forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Requires user to have completed onboarding to access endpoints under this.
// Use after RequireAuth so the current session is already in the request context.
func RequireOnboarded(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := SessionFromContext(r.Context())
		if session == nil {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
			return
		}
		if !isUserOnboarded(session.User) {
			writeError(w, http.StatusForbidden, "ONBOARDING_REQUIRED", "Onboarding required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireExecGroup(checker ExecGroupChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := SessionFromContext((r.Context()))
			if session == nil || session.User == nil {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
				return
			}

			userID, ok := session.User.ID.(string)
			if !ok || userID == "" {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
				return
			}

			hasExecGroup, err := checker.HasExecGroup(r.Context(), userID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to verify executive group")
				return
			}

			if !hasExecGroup {
				writeError(w, http.StatusForbidden, "FORBIDDEN", "Forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isUserOnboarded(user *limen.User) bool {
	if user == nil {
		return false
	}
	completedAt := user.Raw()["onboarding_completed_at"]
	switch v := completedAt.(type) {
	case time.Time:
		return !v.IsZero()
	case *time.Time:
		return v != nil && !v.IsZero()
	default:
		return completedAt != nil
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":    code,
		"message": message,
	})
}

// Extract session from context
func SessionFromContext(ctx context.Context) *limen.ValidatedSession {
	session, _ := ctx.Value(sessionKey).(*limen.ValidatedSession)
	return session
}

func WithRequestMetadata(ctx context.Context) (context.Context, *RequestMetadata) {
	metadata := &RequestMetadata{}
	return context.WithValue(ctx, requestMetadataKey{}, metadata), metadata
}

func RequestMetadataFromContext(ctx context.Context) *RequestMetadata {
	metadata, _ := ctx.Value(requestMetadataKey{}).(*RequestMetadata)
	return metadata
}
