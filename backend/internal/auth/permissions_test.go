package auth

import (
	"context"
	"errors"
	"os"
	"regexp"
	"slices"
	"testing"
)

type fakePermissionLoader struct {
	permissions []string
	err         error
	calls       int
}

func (f *fakePermissionLoader) GetPermissionsForUser(context.Context, string) ([]string, error) {
	f.calls++
	return f.permissions, f.err
}

func TestPermissionsMatchMigrationAndSeed(t *testing.T) {
	for _, path := range []string{
		"../../sql/migrations/20261117000000_seed_permissions.sql",
		"../../sql/seeds/permissions.sql",
	} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		var found []string
		for _, match := range regexp.MustCompile(`\('([a-z_.]+)'\)`).FindAllStringSubmatch(string(content), -1) {
			found = append(found, match[1])
		}

		defined := slices.Clone(AllPermissions)
		slices.Sort(found)
		slices.Sort(defined)
		if !slices.Equal(found, defined) {
			t.Fatalf("permission constants and %s differ:\n file: %v\n code: %v", path, found, defined)
		}
	}
}

func TestBuildPermissionSetAdminHasEveryPermission(t *testing.T) {
	loader := &fakePermissionLoader{}
	permissions, err := buildPermissionSet(context.Background(), loader, "user-id", SystemRoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	if loader.calls != 0 {
		t.Fatal("expected admin to skip the permission lookup")
	}
	for _, permission := range AllPermissions {
		if !permissions.has(permission) {
			t.Fatalf("expected admin to have %s", permission)
		}
	}
}

func TestBuildPermissionSetUsesRolePermissions(t *testing.T) {
	loader := &fakePermissionLoader{permissions: []string{PermissionUsersView, PermissionAuditLogsView}}
	permissions, err := buildPermissionSet(context.Background(), loader, "user-id", "member")
	if err != nil {
		t.Fatal(err)
	}
	if !permissions.has(PermissionUsersView) || !permissions.has(PermissionAuditLogsView) {
		t.Fatal("expected granted permissions to be present")
	}
	if permissions.has(PermissionUsersUpdate) {
		t.Fatal("expected ungranted permission to be absent")
	}
}

func TestBuildPermissionSetReturnsLoaderError(t *testing.T) {
	want := errors.New("database unavailable")
	if _, err := buildPermissionSet(context.Background(), &fakePermissionLoader{err: want}, "user-id", "member"); !errors.Is(err, want) {
		t.Fatalf("expected loader error, got %v", err)
	}
}

func TestHasPermissionWithoutLoadedPermissions(t *testing.T) {
	if HasPermission(context.Background(), PermissionUsersView) {
		t.Fatal("expected no permissions when none were loaded")
	}
}

func TestHasPermissionReadsContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), permissionsKey{}, &permissionSet{
		permissions: map[string]struct{}{PermissionMembershipsDelete: {}},
	})
	if !HasPermission(ctx, PermissionMembershipsDelete) {
		t.Fatal("expected loaded permission to be present")
	}
	if HasPermission(ctx, PermissionUsersUpdate) {
		t.Fatal("expected unloaded permission to be absent")
	}
}

func TestCanAccessUser(t *testing.T) {
	const self = "3c2b1a09-8765-4321-abcd-0123456789ab"
	const other = "8f0f7a4c-1a2b-4c3d-9e5f-6a7b8c9d0e1f"

	for _, tc := range []struct {
		name   string
		actor  string
		role   string
		target string
		want   bool
	}{
		{"admin can see anyone", other, SystemRoleAdmin, self, true},
		{"member can see themself", self, "member", self, true},
		{"member match ignores case and spaces", self, "member", " 3C2B1A09-8765-4321-ABCD-0123456789AB ", true},
		{"member cannot see another user", other, "member", self, false},
		{"missing actor is denied", "", "member", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := canAccessUser(tc.actor, tc.role, tc.target); got != tc.want {
				t.Fatalf("canAccessUser(%q, %q, %q) = %v, want %v", tc.actor, tc.role, tc.target, got, tc.want)
			}
		})
	}
}

func TestCanAccessUserWithoutSession(t *testing.T) {
	if CanAccessUser(context.Background(), "3c2b1a09-8765-4321-abcd-0123456789ab") {
		t.Fatal("expected no access without a session")
	}
}
