package auth

import (
	"context"
	"net/http"
	"strings"
)

// Permission keys granted through roles. Keep in sync with the permissions
// migration and seed; TestPermissionsMatchMigrationAndSeed enforces it.
const (
	PermissionUsersView   = "users.view"
	PermissionUsersUpdate = "users.update"
	PermissionUsersExport = "users.export"

	PermissionMembershipTiersView = "membership_tiers.view"

	PermissionMembershipsView   = "memberships.view"
	PermissionMembershipsCreate = "memberships.create"
	PermissionMembershipsDelete = "memberships.delete"

	PermissionMembershipInvitationsView   = "membership_invitations.view"
	PermissionMembershipInvitationsCreate = "membership_invitations.create"
	PermissionMembershipInvitationsDelete = "membership_invitations.delete"

	PermissionExecProfilesUpdate = "exec_profiles.update"

	PermissionAuditLogsView   = "audit_logs.view"
	PermissionAuditLogsExport = "audit_logs.export"

	PermissionAnalyticsView   = "analytics.view"
	PermissionAnalyticsExport = "analytics.export"

	PermissionLoungeSessionsCreate   = "lounge_sessions.create"
	PermissionLoungeSessionsCheckout = "lounge_sessions.checkout"
	PermissionLoungeSessionsView     = "lounge_sessions.view"
	PermissionLoungeStationsUpdate   = "lounge_stations.update"
)

// AllPermissions lists every permission key.
var AllPermissions = []string{
	PermissionUsersView,
	PermissionUsersUpdate,
	PermissionUsersExport,
	PermissionMembershipTiersView,
	PermissionMembershipsView,
	PermissionMembershipsCreate,
	PermissionMembershipsDelete,
	PermissionMembershipInvitationsView,
	PermissionMembershipInvitationsCreate,
	PermissionMembershipInvitationsDelete,
	PermissionExecProfilesUpdate,
	PermissionAuditLogsView,
	PermissionAuditLogsExport,
	PermissionAnalyticsView,
	PermissionAnalyticsExport,
	PermissionLoungeSessionsCreate,
	PermissionLoungeSessionsCheckout,
	PermissionLoungeSessionsView,
	PermissionLoungeStationsUpdate,
}

type PermissionLoader interface {
	GetPermissionsForUser(ctx context.Context, userID string) ([]string, error)
}

type permissionsKey struct{}

// permissionSet is the caller's permissions for one request. Admins hold all.
type permissionSet struct {
	admin       bool
	permissions map[string]struct{}
}

func (p *permissionSet) has(permission string) bool {
	if p.admin {
		return true
	}
	_, ok := p.permissions[permission]
	return ok
}

// Requires user to hold this permission through any of their roles. Admins
// bypass the check. Permissions are loaded once per request and cached in the
// context for HasPermission. Use after RequireAuth.
// eg. r.With(auth.RequirePermission(roleRepository, auth.PermissionUsersView)).Get(...)
func RequirePermission(loader PermissionLoader, permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, permissions, ok := loadPermissions(w, r, loader)
			if !ok {
				return
			}
			if !permissions.has(permission) {
				writeError(w, http.StatusForbidden, "FORBIDDEN", "Forbidden")
				return
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// HasPermission reports whether the caller holds the permission. It returns
// false unless RequirePermission ran earlier in the chain.
func HasPermission(ctx context.Context, permission string) bool {
	permissions, _ := ctx.Value(permissionsKey{}).(*permissionSet)
	return permissions != nil && permissions.has(permission)
}

// CanAccessUser reports whether the caller may see data that belongs to the
// target user: admins can see anyone, everyone else only themself.
func CanAccessUser(ctx context.Context, targetUserID string) bool {
	session := SessionFromContext(ctx)
	if session == nil || session.User == nil {
		return false
	}
	actorUserID, _ := session.User.ID.(string)
	return canAccessUser(actorUserID, systemRole(session.User), targetUserID)
}

func canAccessUser(actorUserID string, actorSystemRole string, targetUserID string) bool {
	if actorSystemRole == SystemRoleAdmin {
		return true
	}
	actorUserID = strings.TrimSpace(actorUserID)
	return actorUserID != "" && strings.EqualFold(actorUserID, strings.TrimSpace(targetUserID))
}

// IsSystemAdmin reports whether the caller is a system admin.
func IsSystemAdmin(ctx context.Context) bool {
	session := SessionFromContext(ctx)
	return session != nil && systemRole(session.User) == SystemRoleAdmin
}

// loadPermissions returns the caller's permissions, reusing the cached set. It
// writes the error response and returns ok=false on failure.
func loadPermissions(
	w http.ResponseWriter,
	r *http.Request,
	loader PermissionLoader,
) (context.Context, *permissionSet, bool) {
	ctx := r.Context()
	if cached, ok := ctx.Value(permissionsKey{}).(*permissionSet); ok {
		return ctx, cached, true
	}

	session := SessionFromContext(ctx)
	if session == nil || session.User == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return ctx, nil, false
	}
	userID, ok := session.User.ID.(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return ctx, nil, false
	}

	permissions, err := buildPermissionSet(ctx, loader, userID, systemRole(session.User))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to verify permissions")
		return ctx, nil, false
	}
	return context.WithValue(ctx, permissionsKey{}, permissions), permissions, true
}

// buildPermissionSet skips the lookup for admins.
func buildPermissionSet(
	ctx context.Context,
	loader PermissionLoader,
	userID string,
	systemRole string,
) (*permissionSet, error) {
	if systemRole == SystemRoleAdmin {
		return &permissionSet{admin: true}, nil
	}

	keys, err := loader.GetPermissionsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	permissions := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		permissions[key] = struct{}{}
	}
	return &permissionSet{permissions: permissions}, nil
}
