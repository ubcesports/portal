package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/repository"
)

const (
	testRoleID    = "11111111-2222-4333-8444-555555555555"
	testMissingID = "99999999-8888-4777-8666-555555555555"
	testRoleName  = "Treasurer"
	testAssignee  = "Alex Day"
)

// fakeRoleStore is an in-memory repository.RoleStore for the assign and
// unassign tests. WithTx just runs the callback against the same store.
type fakeRoleStore struct {
	roles       map[string]string // role ID -> name
	users       map[string]string // user ID -> full name
	assignments map[string]bool   // role ID + user ID
	assignedBy  []string

	assignErr   error
	unassignErr error
	auditErr    error

	auditLogs []repository.RoleAuditLogParams
}

var _ repository.RoleStore = (*fakeRoleStore)(nil)

func newFakeRoleStore() *fakeRoleStore {
	return &fakeRoleStore{
		roles:       map[string]string{testRoleID: testRoleName},
		users:       map[string]string{testTargetID: testAssignee},
		assignments: map[string]bool{},
	}
}

func (f *fakeRoleStore) key(roleID, userID string) string { return roleID + "|" + userID }

func (f *fakeRoleStore) GetByID(_ context.Context, id string) (*dto.RoleDetailsResponse, error) {
	name, ok := f.roles[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return &dto.RoleDetailsResponse{RoleResponse: dto.RoleResponse{ID: id, Name: name}}, nil
}

func (f *fakeRoleStore) GetAssigneeName(_ context.Context, userID string) (string, error) {
	name, ok := f.users[userID]
	if !ok {
		return "", pgx.ErrNoRows
	}
	return name, nil
}

func (f *fakeRoleStore) AssignUser(_ context.Context, roleID, userID, assignedByUserID string) (bool, error) {
	if f.assignErr != nil {
		return false, f.assignErr
	}
	k := f.key(roleID, userID)
	if f.assignments[k] {
		return false, nil
	}
	f.assignments[k] = true
	f.assignedBy = append(f.assignedBy, assignedByUserID)
	return true, nil
}

func (f *fakeRoleStore) UnassignUser(_ context.Context, roleID, userID string) (bool, error) {
	if f.unassignErr != nil {
		return false, f.unassignErr
	}
	k := f.key(roleID, userID)
	if !f.assignments[k] {
		return false, nil
	}
	delete(f.assignments, k)
	return true, nil
}

func (f *fakeRoleStore) CreateAuditLog(_ context.Context, params repository.RoleAuditLogParams) error {
	if f.auditErr != nil {
		return f.auditErr
	}
	f.auditLogs = append(f.auditLogs, params)
	return nil
}

func (f *fakeRoleStore) WithTx(_ context.Context, fn func(repository.RoleStore) error) error {
	return fn(f)
}

func (f *fakeRoleStore) GetAll(context.Context) ([]dto.RoleResponse, error) { return nil, nil }
func (f *fakeRoleStore) GetForUser(context.Context, string) ([]dto.UserRolePermissions, error) {
	return nil, nil
}
func (f *fakeRoleStore) Create(context.Context, dto.CreateRoleRequest, string) (string, error) {
	return "", nil
}
func (f *fakeRoleStore) Update(context.Context, string, dto.UpdateRoleRequest) error { return nil }
func (f *fakeRoleStore) Delete(context.Context, string) error                        { return nil }

func assertSingleRoleAudit(
	t *testing.T,
	store *fakeRoleStore,
	action string,
	outcome db.AdminAuditOutcomeType,
	description string,
) {
	t.Helper()
	if len(store.auditLogs) != 1 {
		t.Fatalf("expected exactly one audit log, got %#v", store.auditLogs)
	}
	got := store.auditLogs[0]
	if got.Action != action || got.Outcome != outcome {
		t.Fatalf("expected %s/%s audit, got %s/%s", action, outcome, got.Action, got.Outcome)
	}
	if got.ActorUserID != testActorID || got.RequestID != testRequestID {
		t.Fatalf("audit log lost the actor or request ID: %#v", got)
	}
	if !strings.Contains(got.Description, description) {
		t.Fatalf("expected audit description to contain %q, got %q", description, got.Description)
	}
}

func assignUser(store *fakeRoleStore, roleID, userID string) error {
	svc := &RoleService{repository: store}
	return svc.AssignUser(context.Background(), roleID, dto.AssignRoleRequest{UserID: userID}, testActorID, testRequestID)
}

func unassignUser(store *fakeRoleStore, roleID, userID string) error {
	svc := &RoleService{repository: store}
	return svc.UnassignUser(context.Background(), roleID, userID, testActorID, testRequestID)
}

func TestAssignUser(t *testing.T) {
	store := newFakeRoleStore()

	if err := assignUser(store, testRoleID, testTargetID); err != nil {
		t.Fatal(err)
	}
	if !store.assignments[store.key(testRoleID, testTargetID)] {
		t.Fatal("expected the role to be assigned")
	}
	if len(store.assignedBy) != 1 || store.assignedBy[0] != testActorID {
		t.Fatalf("expected the actor recorded as assigner, got %v", store.assignedBy)
	}
	assertSingleRoleAudit(t, store, actionRoleAssigned, db.AdminAuditOutcomeTypeSuccess, "Assigned role Treasurer to Alex Day")
}

func TestAssignUserTrimsUserID(t *testing.T) {
	store := newFakeRoleStore()

	if err := assignUser(store, testRoleID, "  "+testTargetID+"\n"); err != nil {
		t.Fatal(err)
	}
	if !store.assignments[store.key(testRoleID, testTargetID)] {
		t.Fatal("expected the trimmed user ID to be assigned")
	}
}

func TestAssignUserRejectsMalformedIDs(t *testing.T) {
	for _, tc := range []struct{ name, roleID, userID string }{
		{"role", "not-a-uuid", testTargetID},
		{"user", testRoleID, "not-a-uuid"},
		{"empty user", testRoleID, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := newFakeRoleStore()
			err := assignUser(store, tc.roleID, tc.userID)
			if !errors.Is(err, ErrInvalidRole) {
				t.Fatalf("expected ErrInvalidRole, got %v", err)
			}
			if len(store.assignments) != 0 {
				t.Fatal("nothing should be assigned")
			}
			assertSingleRoleAudit(t, store, actionRoleAssigned, db.AdminAuditOutcomeTypeFailed, "Failed to assign role")
		})
	}
}

func TestAssignUserRoleNotFound(t *testing.T) {
	store := newFakeRoleStore()

	err := assignUser(store, testMissingID, testTargetID)
	if !errors.Is(err, ErrRoleNotFound) {
		t.Fatalf("expected ErrRoleNotFound, got %v", err)
	}
	assertSingleRoleAudit(t, store, actionRoleAssigned, db.AdminAuditOutcomeTypeFailed, "Failed to assign role "+testMissingID)
}

func TestAssignUserUserNotFound(t *testing.T) {
	store := newFakeRoleStore()

	err := assignUser(store, testRoleID, testMissingID)
	if !errors.Is(err, ErrRoleUserNotFound) {
		t.Fatalf("expected ErrRoleUserNotFound, got %v", err)
	}
	if len(store.assignments) != 0 {
		t.Fatal("nothing should be assigned")
	}
	assertSingleRoleAudit(t, store, actionRoleAssigned, db.AdminAuditOutcomeTypeFailed, "Failed to assign role Treasurer")
}

func TestAssignUserAlreadyAssigned(t *testing.T) {
	store := newFakeRoleStore()
	store.assignments[store.key(testRoleID, testTargetID)] = true

	err := assignUser(store, testRoleID, testTargetID)
	if !errors.Is(err, ErrRoleAlreadyAssigned) {
		t.Fatalf("expected ErrRoleAlreadyAssigned, got %v", err)
	}
	assertSingleRoleAudit(t, store, actionRoleAssigned, db.AdminAuditOutcomeTypeFailed, "Failed to assign role Treasurer to Alex Day")
}

func TestAssignUserDatabaseFailure(t *testing.T) {
	store := newFakeRoleStore()
	store.assignErr = errors.New("database failure")

	err := assignUser(store, testRoleID, testTargetID)
	if !errors.Is(err, store.assignErr) {
		t.Fatalf("expected the database error, got %v", err)
	}
	assertSingleRoleAudit(t, store, actionRoleAssigned, db.AdminAuditOutcomeTypeFailed, "Failed to assign role Treasurer to Alex Day")
}

func TestAssignUserReturnsAuditFailure(t *testing.T) {
	store := newFakeRoleStore()
	store.auditErr = errors.New("audit failure")

	err := assignUser(store, testRoleID, testTargetID)
	if !errors.Is(err, store.auditErr) {
		t.Fatalf("expected the audit error, got %v", err)
	}
}

func TestUnassignUser(t *testing.T) {
	store := newFakeRoleStore()
	store.assignments[store.key(testRoleID, testTargetID)] = true

	if err := unassignUser(store, testRoleID, testTargetID); err != nil {
		t.Fatal(err)
	}
	if store.assignments[store.key(testRoleID, testTargetID)] {
		t.Fatal("expected the role to be removed")
	}
	assertSingleRoleAudit(t, store, actionRoleUnassigned, db.AdminAuditOutcomeTypeSuccess, "Removed role Treasurer from Alex Day")
}

func TestUnassignUserRejectsMalformedIDs(t *testing.T) {
	for _, tc := range []struct{ name, roleID, userID string }{
		{"role", "not-a-uuid", testTargetID},
		{"user", testRoleID, "not-a-uuid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := newFakeRoleStore()
			store.assignments[store.key(testRoleID, testTargetID)] = true

			err := unassignUser(store, tc.roleID, tc.userID)
			if !errors.Is(err, ErrInvalidRole) {
				t.Fatalf("expected ErrInvalidRole, got %v", err)
			}
			if !store.assignments[store.key(testRoleID, testTargetID)] {
				t.Fatal("the existing assignment must be untouched")
			}
			assertSingleRoleAudit(t, store, actionRoleUnassigned, db.AdminAuditOutcomeTypeFailed, "Failed to remove role")
		})
	}
}

func TestUnassignUserRoleNotFound(t *testing.T) {
	store := newFakeRoleStore()

	err := unassignUser(store, testMissingID, testTargetID)
	if !errors.Is(err, ErrRoleNotFound) {
		t.Fatalf("expected ErrRoleNotFound, got %v", err)
	}
	assertSingleRoleAudit(t, store, actionRoleUnassigned, db.AdminAuditOutcomeTypeFailed, "Failed to remove role "+testMissingID)
}

func TestUnassignUserUserNotFound(t *testing.T) {
	store := newFakeRoleStore()

	err := unassignUser(store, testRoleID, testMissingID)
	if !errors.Is(err, ErrRoleUserNotFound) {
		t.Fatalf("expected ErrRoleUserNotFound, got %v", err)
	}
	assertSingleRoleAudit(t, store, actionRoleUnassigned, db.AdminAuditOutcomeTypeFailed, "Failed to remove role Treasurer")
}

func TestUnassignUserNotAssigned(t *testing.T) {
	store := newFakeRoleStore()

	err := unassignUser(store, testRoleID, testTargetID)
	if !errors.Is(err, ErrRoleNotAssigned) {
		t.Fatalf("expected ErrRoleNotAssigned, got %v", err)
	}
	assertSingleRoleAudit(t, store, actionRoleUnassigned, db.AdminAuditOutcomeTypeFailed, "Failed to remove role Treasurer from Alex Day")
}

func TestUnassignUserDatabaseFailure(t *testing.T) {
	store := newFakeRoleStore()
	store.assignments[store.key(testRoleID, testTargetID)] = true
	store.unassignErr = errors.New("database failure")

	err := unassignUser(store, testRoleID, testTargetID)
	if !errors.Is(err, store.unassignErr) {
		t.Fatalf("expected the database error, got %v", err)
	}
	assertSingleRoleAudit(t, store, actionRoleUnassigned, db.AdminAuditOutcomeTypeFailed, "Failed to remove role Treasurer from Alex Day")
}

func TestAssignThenUnassignRoundTrip(t *testing.T) {
	store := newFakeRoleStore()

	if err := assignUser(store, testRoleID, testTargetID); err != nil {
		t.Fatal(err)
	}
	if err := assignUser(store, testRoleID, testTargetID); !errors.Is(err, ErrRoleAlreadyAssigned) {
		t.Fatalf("expected a second assign to conflict, got %v", err)
	}
	if err := unassignUser(store, testRoleID, testTargetID); err != nil {
		t.Fatal(err)
	}
	if err := unassignUser(store, testRoleID, testTargetID); !errors.Is(err, ErrRoleNotAssigned) {
		t.Fatalf("expected a second unassign to report not assigned, got %v", err)
	}
}
