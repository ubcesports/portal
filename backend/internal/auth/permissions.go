package auth

import (
	"context"
	"net/http"
)

// Permission keys granted through roles. These must match the rows seeded in
// sql/seeds/permissions.sql; TestPermissionsMatchSeed enforces that.
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

	PermissionRolesView   = "roles.view"
	PermissionRolesCreate = "roles.create"
	PermissionRolesUpdate = "roles.update"
	PermissionRolesDelete = "roles.delete"
)

// AllPermissions lists every permission key defined above.
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
	PermissionRolesView,
	PermissionRolesCreate,
	PermissionRolesUpdate,
	PermissionRolesDelete,
}

type PermissionLoader interface {
	GetPermissionsForUser(ctx context.Context, userID string) ([]string, error)
}

type permissionsKey struct{}

// permissionSet is the caller's effective permissions for one request.
// System admins hold every permission.
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

// Requires user to hold this permission, through any of their roles, to access
// endpoints under this. Users with the "admin" system role bypass the check.
// The caller's permissions are loaded once per request and cached in the
// context, so handlers can check further permissions with HasPermission.
// Use after RequireAuth.
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

// HasPermission reports whether the caller holds the permission. It only
// sees permissions loaded by RequirePermission earlier in the chain and
// returns false when none were loaded.
func HasPermission(ctx context.Context, permission string) bool {
	permissions, _ := ctx.Value(permissionsKey{}).(*permissionSet)
	return permissions != nil && permissions.has(permission)
}

// IsSystemAdmin reports whether the caller has the "admin" system role.
func IsSystemAdmin(ctx context.Context) bool {
	session := SessionFromContext(ctx)
	return session != nil && systemRole(session.User) == SystemRoleAdmin
}

// loadPermissions returns the caller's permissions, reusing ones already
// cached in the context. It writes an error response and returns ok=false
// when the caller is unauthenticated or the lookup fails.
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

// buildPermissionSet skips the lookup for system admins, who hold every
// permission.
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
