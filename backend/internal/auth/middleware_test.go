package auth

import (
	"context"
	"errors"
	"slices"
	"testing"
)

type fakeRoleChecker struct {
	roles  []string
	err    error
	called bool
}

func (f *fakeRoleChecker) HasAnyRole(_ context.Context, _ string, roleNames []string) (bool, error) {
	f.called = true
	if f.err != nil {
		return false, f.err
	}
	for _, name := range roleNames {
		if slices.Contains(f.roles, name) {
			return true, nil
		}
	}
	return false, nil
}

func TestHasRequiredRoleAdminBypassesCheck(t *testing.T) {
	checker := &fakeRoleChecker{}
	allowed, err := hasRequiredRole(context.Background(), checker, "user-id", SystemRoleAdmin, []string{"Treasurer"})
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("expected admin to be allowed")
	}
	if checker.called {
		t.Fatal("expected admin to bypass the role lookup")
	}
}

func TestHasRequiredRoleChecksAssignedRoles(t *testing.T) {
	tests := []struct {
		name     string
		assigned []string
		required []string
		want     bool
	}{
		{"has required role", []string{"Treasurer"}, []string{"Treasurer"}, true},
		{"has one of several", []string{"President"}, []string{"Treasurer", "President"}, true},
		{"missing role", []string{"Events"}, []string{"Treasurer"}, false},
		{"no roles", nil, []string{"Treasurer"}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			checker := &fakeRoleChecker{roles: test.assigned}
			allowed, err := hasRequiredRole(context.Background(), checker, "user-id", "member", test.required)
			if err != nil {
				t.Fatal(err)
			}
			if allowed != test.want {
				t.Fatalf("expected %v, got %v", test.want, allowed)
			}
		})
	}
}

func TestHasRequiredRoleDeniesWhenNoRolesGiven(t *testing.T) {
	checker := &fakeRoleChecker{}
	allowed, err := hasRequiredRole(context.Background(), checker, "user-id", "member", nil)
	if err != nil {
		t.Fatal(err)
	}
	if allowed || checker.called {
		t.Fatalf("expected non-admin to be denied without a lookup, allowed=%v called=%v", allowed, checker.called)
	}
}

func TestHasRequiredRoleReturnsCheckerError(t *testing.T) {
	want := errors.New("database unavailable")
	checker := &fakeRoleChecker{err: want}
	if _, err := hasRequiredRole(context.Background(), checker, "user-id", "member", []string{"Treasurer"}); !errors.Is(err, want) {
		t.Fatalf("expected checker error, got %v", err)
	}
}
