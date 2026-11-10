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
type RoleChecker interface {
	HasAnyRole(ctx context.Context, userID string, roleNames []string) (bool, error)
}

const sessionKey contextKey = "session"

// SystemRoleAdmin is the users.role value that grants full access, including
// bypassing RequireRole checks.
const SystemRoleAdmin = "admin"

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

// Requires user to have one of these system roles (the users.role column) to
// access endpoints under this.
// eg. r.Use(auth.RequireSystem("admin")) or r.Use(auth.RequireSystem("admin", "member"))
func RequireSystem(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := SessionFromContext(r.Context())
			if session == nil || session.User == nil {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
				return
			}
			if !allowed[systemRole(session.User)] {
				writeError(w, http.StatusForbidden, "FORBIDDEN", "Forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Requires user to be assigned one of these custom roles (from the roles
// system) to access endpoints under this. Users with the "admin" system role
// bypass the check. Use after RequireAuth.
// eg. r.Use(auth.RequireRole(roleRepository, "Treasurer")) or
// r.Use(auth.RequireRole(roleRepository, "Treasurer", "President"))
func RequireRole(checker RoleChecker, roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := SessionFromContext(r.Context())
			if session == nil || session.User == nil {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
				return
			}

			userID, ok := session.User.ID.(string)
			if !ok || userID == "" {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
				return
			}

			allowed, err := hasRequiredRole(r.Context(), checker, userID, systemRole(session.User), roles)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to verify role")
				return
			}
			if !allowed {
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

// hasRequiredRole lets system admins through and otherwise checks whether the
// user is assigned any of the required custom roles.
func hasRequiredRole(
	ctx context.Context,
	checker RoleChecker,
	userID string,
	systemRole string,
	roles []string,
) (bool, error) {
	if systemRole == SystemRoleAdmin {
		return true, nil
	}
	if len(roles) == 0 {
		return false, nil
	}
	return checker.HasAnyRole(ctx, userID, roles)
}

func systemRole(user *limen.User) string {
	if user == nil {
		return ""
	}
	role, _ := user.Raw()["role"].(string)
	return role
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
