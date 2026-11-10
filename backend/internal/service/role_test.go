package service

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ubcesports/memberships/internal/dto"
)

func TestNormalizeCreateRole(t *testing.T) {
	request, err := normalizeCreateRole(dto.CreateRoleRequest{
		Name:           "  Treasurer ",
		Description:    ptr("   "),
		Color:          ptr(" #a1b2c3 "),
		DisplayOrder:   ptr(int32(2)),
		PermissionKeys: []string{" users.view ", "memberships.view", "users.view"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if request.Name != "Treasurer" {
		t.Fatalf("unexpected name: %q", request.Name)
	}
	if request.Description != nil {
		t.Fatalf("expected blank description to be nil, got %q", *request.Description)
	}
	if request.Color == nil || *request.Color != "#A1B2C3" {
		t.Fatalf("unexpected color: %v", request.Color)
	}
	if strings.Join(request.PermissionKeys, ",") != "users.view,memberships.view" {
		t.Fatalf("unexpected permission keys: %v", request.PermissionKeys)
	}
}

func TestNormalizeCreateRoleRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name    string
		request dto.CreateRoleRequest
		want    error
	}{
		{"blank name", dto.CreateRoleRequest{Name: "   "}, ErrInvalidRole},
		{"long name", dto.CreateRoleRequest{Name: strings.Repeat("a", maxRoleNameLength+1)}, ErrInvalidRole},
		{"bad color", dto.CreateRoleRequest{Name: "Exec", Color: ptr("red")}, ErrInvalidRole},
		{"short color", dto.CreateRoleRequest{Name: "Exec", Color: ptr("#fff")}, ErrInvalidRole},
		{"negative order", dto.CreateRoleRequest{Name: "Exec", DisplayOrder: ptr(int32(-1))}, ErrInvalidRole},
		{"blank permission", dto.CreateRoleRequest{Name: "Exec", PermissionKeys: []string{" "}}, ErrInvalidRolePermission},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := normalizeCreateRole(test.request); !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}

func TestNormalizeCreateRoleAllowsMaxLengthMultibyteName(t *testing.T) {
	name := strings.Repeat("é", maxRoleNameLength)
	if _, err := normalizeCreateRole(dto.CreateRoleRequest{Name: name}); err != nil {
		t.Fatalf("expected %d-character name to be valid, got %v", maxRoleNameLength, err)
	}
}

func TestNormalizeUpdateRole(t *testing.T) {
	keys := []string{"roles.view", " roles.view "}
	request, err := normalizeUpdateRole(dto.UpdateRoleRequest{
		Name:           ptr(" Exec "),
		Color:          ptr("#abcdef"),
		PermissionKeys: &keys,
	})
	if err != nil {
		t.Fatal(err)
	}
	if *request.Name != "Exec" || *request.Color != "#ABCDEF" {
		t.Fatalf("unexpected normalized request: %#v", request)
	}
	if strings.Join(*request.PermissionKeys, ",") != "roles.view" {
		t.Fatalf("unexpected permission keys: %v", *request.PermissionKeys)
	}
}

func TestNormalizeUpdateRoleAllowsClearingPermissions(t *testing.T) {
	request, err := normalizeUpdateRole(dto.UpdateRoleRequest{PermissionKeys: &[]string{}})
	if err != nil {
		t.Fatal(err)
	}
	if request.PermissionKeys == nil || len(*request.PermissionKeys) != 0 {
		t.Fatalf("expected empty permission keys, got %v", request.PermissionKeys)
	}
}

func TestNormalizeUpdateRoleRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name    string
		request dto.UpdateRoleRequest
	}{
		{"empty update", dto.UpdateRoleRequest{}},
		{"blank name", dto.UpdateRoleRequest{Name: ptr(" ")}},
		{"bad color", dto.UpdateRoleRequest{Color: ptr("#12345G")}},
		{"negative order", dto.UpdateRoleRequest{DisplayOrder: ptr(int32(-5))}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := normalizeUpdateRole(test.request); !errors.Is(err, ErrInvalidRole) {
				t.Fatalf("expected invalid role error, got %v", err)
			}
		})
	}
}

func TestRoleError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{"not found", fmt.Errorf("get role by ID: %w", pgx.ErrNoRows), ErrRoleNotFound},
		{"duplicate name", fmt.Errorf("create role: %w", &pgconn.PgError{Code: "23505"}), ErrRoleNameTaken},
		{"unknown permission", fmt.Errorf("add role permission keys: %w", &pgconn.PgError{Code: "23503"}), ErrInvalidRolePermission},
		{"check violation", &pgconn.PgError{Code: "23514"}, ErrInvalidRole},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := roleError(test.err); !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}

	other := errors.New("connection reset")
	if err := roleError(other); err != other {
		t.Fatalf("expected unmapped error to pass through, got %v", err)
	}
}

func TestDescribeRoleUpdate(t *testing.T) {
	before := &dto.RoleDetailsResponse{
		RoleResponse:   dto.RoleResponse{Name: "Exec", DisplayOrder: 1},
		PermissionKeys: []string{"users.view"},
	}
	after := &dto.RoleDetailsResponse{
		RoleResponse:   dto.RoleResponse{Name: "Executive", Color: ptr("#FFFFFF"), DisplayOrder: 1},
		PermissionKeys: []string{"roles.view", "users.view"},
	}

	got := describeRoleUpdate(before, after)
	want := `Updated role Executive: name "Exec" → "Executive"; color none → #FFFFFF; permissions users.view → roles.view, users.view`
	if got != want {
		t.Fatalf("unexpected description:\n got: %s\nwant: %s", got, want)
	}
}
