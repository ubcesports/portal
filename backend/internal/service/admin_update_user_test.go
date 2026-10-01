package service

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/repository"
)

const (
	testActorID      = "8f0f7a4c-1a2b-4c3d-9e5f-6a7b8c9d0e1f"
	testTargetID     = "3c2b1a09-8765-4321-abcd-0123456789ab"
	testMembershipID = "2d746a56-c977-49e0-a04c-20504cdb07c0"
	testRequestID    = "req-1"
)

var nonStudentIDRegex = regexp.MustCompile(`^N\d{7}$`)

type studentInfoUpdate struct {
	isStudent bool
	studentID string
}

// fakeAdminStore is an in-memory repository.AdminStore for exercising the admin
// service without a database.
type fakeAdminStore struct {
	fullNameUpdates   []string
	updateFullNameErr error
	user              db.GetAdminUserByIDRow
	getUserErr        error

	takenStudentIDs    map[string]bool
	allStudentIDsTaken bool

	memberships              []db.GetAllMembershipsWithTransactionsRow
	hasActiveMembership      bool
	updateStudentInfoErr     error
	studentInfoUpdates       []studentInfoUpdate
	roleUpdates              []db.RoleType
	addedGroups              []db.GroupType
	removedGroups            []db.GroupType
	cancelledMembershipUsers []string
	cancelledMembershipIDs   []string
	cancelMembershipResult   bool
	cancelMembershipErr      error

	auditLogs []db.CreateAdminAuditLogParams
}

var _ repository.AdminStore = (*fakeAdminStore)(nil)

func (f *fakeAdminStore) UpdateUserFullName(_ context.Context, _ string, fullName string) error {
	if f.updateFullNameErr != nil {
		return f.updateFullNameErr
	}
	f.fullNameUpdates = append(f.fullNameUpdates, fullName)
	f.user.FullName = fullName
	return nil
}

func TestUpdateUserFullName(t *testing.T) {
	store := newFakeAdminStore(t, false, "N1234567", db.RoleTypeMember, "member")
	svc := &AdminService{adminRepository: store}
	profile, err := svc.UpdateUser(context.Background(), testActorID, testTargetID, testRequestID, UpdateUserRequest{FullName: ptr("  Renée O'Connor  ")})
	if err != nil {
		t.Fatal(err)
	}
	if profile.FullName != "Renée O'Connor" || len(store.fullNameUpdates) != 1 {
		t.Fatalf("expected trimmed name in updated profile, got %#v", profile)
	}
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeSuccess, actionFullNameUpdated)
	if store.auditLogs[0].Description.String != "Updated full name from \"Sudi Mango\" to \"Renée O'Connor\"" {
		t.Fatalf("unexpected audit description: %v", store.auditLogs[0].Description)
	}
}

func TestUpdateUserFullNameValidationAndNoOp(t *testing.T) {
	for _, tc := range []struct {
		name    string
		value   *string
		invalid bool
	}{
		{"omitted", nil, false}, {"unchanged", ptr(" Sudi Mango "), false},
		{"empty", ptr(""), true}, {"whitespace", ptr(" \t\n\u00a0"), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := newFakeAdminStore(t, false, "N1234567", db.RoleTypeMember, "member")
			err := updateUser(t, store, UpdateUserRequest{FullName: tc.value})
			if tc.invalid {
				if !errors.Is(err, ErrValidation) {
					t.Fatalf("expected validation error, got %v", err)
				}
				assertAuditActions(t, store, db.AdminAuditOutcomeTypeFailed, actionFullNameUpdated)
			} else if err != nil || len(store.auditLogs) != 0 {
				t.Fatalf("expected no-op, got error %v and logs %v", err, store.auditLogs)
			}
			if len(store.fullNameUpdates) != 0 {
				t.Fatal("unexpected name update")
			}
		})
	}
}

func TestUpdateUserFullNamePersistenceFailure(t *testing.T) {
	store := newFakeAdminStore(t, false, "N1234567", db.RoleTypeMember, "member")
	store.updateFullNameErr = errors.New("database failure")
	err := updateUser(t, store, UpdateUserRequest{FullName: ptr("New Name")})
	if !errors.Is(err, store.updateFullNameErr) {
		t.Fatalf("expected database error, got %v", err)
	}
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeFailed, actionFullNameUpdated)
}

func (f *fakeAdminStore) GetUsers(context.Context, db.GetUsersAdminParams) ([]db.GetUsersAdminRow, error) {
	return nil, nil
}

func (f *fakeAdminStore) CountUsers(context.Context, db.CountUsersAdminParams) (int64, error) {
	return 0, nil
}

func (f *fakeAdminStore) GetAdminMembershipTierOptions(context.Context) ([]db.GetAdminMembershipTierOptionsRow, error) {
	return nil, nil
}

func (f *fakeAdminStore) CreateAdminAuditLog(_ context.Context, params db.CreateAdminAuditLogParams) error {
	f.auditLogs = append(f.auditLogs, params)
	return nil
}

func (f *fakeAdminStore) CountAdminAuditLogs(context.Context, pgtype.Text) (int64, error) {
	return 0, nil
}

func (f *fakeAdminStore) GetAdminAuditLogs(context.Context, db.GetAdminAuditLogsParams) ([]db.GetAdminAuditLogsRow, error) {
	return nil, nil
}

func (f *fakeAdminStore) GetUserByID(context.Context, string) (db.GetAdminUserByIDRow, error) {
	if f.getUserErr != nil {
		return db.GetAdminUserByIDRow{}, f.getUserErr
	}
	return f.user, nil
}

func (f *fakeAdminStore) UpdateUserStudentInfo(_ context.Context, _ string, isStudent bool, studentId string) error {
	if f.updateStudentInfoErr != nil {
		return f.updateStudentInfoErr
	}
	f.studentInfoUpdates = append(f.studentInfoUpdates, studentInfoUpdate{isStudent: isStudent, studentID: studentId})
	return nil
}

func (f *fakeAdminStore) StudentIDExists(_ context.Context, studentId string) (bool, error) {
	return f.allStudentIDsTaken || f.takenStudentIDs[studentId], nil
}

func (f *fakeAdminStore) UpdateUserRole(_ context.Context, _ string, role db.RoleType) error {
	f.roleUpdates = append(f.roleUpdates, role)
	return nil
}

func (f *fakeAdminStore) AddUserGroup(_ context.Context, _ string, group db.GroupType) error {
	f.addedGroups = append(f.addedGroups, group)
	return nil
}

func (f *fakeAdminStore) CreateExecProfile(ctx context.Context, userId string, title pgtype.Text, displayOrder pgtype.Int4, displayGroup db.NullExecDisplayGroupType) error {
	return nil
}
func (f *fakeAdminStore) HasExecProfile(ctx context.Context, userId string) (bool, error) {
	return false, nil
}
func (f *fakeAdminStore) HasExecGroup(ctx context.Context, userId string) (bool, error) {
	return false, nil
}

func (f *fakeAdminStore) RemoveExecProfile(_ context.Context, _ string) error {
	return nil
}

func (f *fakeAdminStore) UpdateExecProfile(_ context.Context, _ string, title pgtype.Text, displayOrder pgtype.Int4, displayGroup db.NullExecDisplayGroupType) (db.GetExecProfileByUserIDRow, error) {
	return db.GetExecProfileByUserIDRow{}, nil
}

func (f *fakeAdminStore) RemoveUserGroup(_ context.Context, _ string, group db.GroupType) error {
	f.removedGroups = append(f.removedGroups, group)
	return nil
}

func (f *fakeAdminStore) GetUserMemberships(context.Context, string) ([]db.GetAllMembershipsWithTransactionsRow, error) {
	return f.memberships, nil
}

func (f *fakeAdminStore) HasActiveMembership(context.Context, string) (bool, error) {
	return f.hasActiveMembership, nil
}

func (f *fakeAdminStore) CancelActiveMembershipByUserIdAndMembershipId(
	_ context.Context,
	userId string,
	membershipId string,
	_ time.Time,
) (bool, error) {
	if f.cancelMembershipErr != nil {
		return false, f.cancelMembershipErr
	}

	f.cancelledMembershipUsers = append(f.cancelledMembershipUsers, userId)
	f.cancelledMembershipIDs = append(f.cancelledMembershipIDs, membershipId)
	return f.cancelMembershipResult, nil
}

func (f *fakeAdminStore) WithTx(ctx context.Context, fn func(repository.AdminStore) error) error {
	return fn(f)
}

/*
	Helpers
*/

func newFakeAdminStore(t *testing.T, isStudent bool, studentID string, role db.RoleType, groups ...string) *fakeAdminStore {
	t.Helper()

	var userID pgtype.UUID
	if err := userID.Scan(testTargetID); err != nil {
		t.Fatalf("unable to build test user ID: %v", err)
	}

	return &fakeAdminStore{
		user: db.GetAdminUserByIDRow{
			ID:        userID,
			Email:     "sudi@example.com",
			StudentID: pgtype.Text{String: studentID, Valid: studentID != ""},
			Role:      role,
			FullName:  "Sudi Mango",
			IsStudent: isStudent,
			Groups:    groups,
		},
		takenStudentIDs: map[string]bool{},
	}
}

func updateUser(t *testing.T, store *fakeAdminStore, req UpdateUserRequest) error {
	t.Helper()

	adminService := &AdminService{adminRepository: store}
	_, err := adminService.UpdateUser(context.Background(), testActorID, testTargetID, testRequestID, req)
	return err
}

// auditActions returns the recorded audit entries that carry the given outcome.
func auditActions(store *fakeAdminStore, outcome db.AdminAuditOutcomeType) []string {
	actions := make([]string, 0, len(store.auditLogs))
	for _, log := range store.auditLogs {
		if log.Outcome == outcome {
			actions = append(actions, log.Action)
		}
	}
	return actions
}

func assertAuditActions(t *testing.T, store *fakeAdminStore, outcome db.AdminAuditOutcomeType, expected ...string) {
	t.Helper()

	actions := auditActions(store, outcome)
	if len(actions) != len(expected) {
		t.Fatalf("expected %v audit actions with outcome %q, got %v", expected, outcome, actions)
	}
	for i, action := range expected {
		if actions[i] != action {
			t.Fatalf("expected audit action %q at index %d, got %v", action, i, actions)
		}
	}
}

func ptr[T any](value T) *T {
	return &value
}

/*
	Student ID
*/

func TestUpdateUserRejectsStudentIDForNonStudent(t *testing.T) {
	store := newFakeAdminStore(t, false, "N1234567", db.RoleTypeMember, "member")

	err := updateUser(t, store, UpdateUserRequest{StudentID: ptr("12345678")})

	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(store.studentInfoUpdates) != 0 {
		t.Fatalf("expected no student info update, got %#v", store.studentInfoUpdates)
	}
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeFailed, actionStudentIDUpdated)
}

func TestUpdateUserAcceptsStudentIDForStudent(t *testing.T) {
	store := newFakeAdminStore(t, true, "12345678", db.RoleTypeMember, "member")

	if err := updateUser(t, store, UpdateUserRequest{StudentID: ptr("87654321")}); err != nil {
		t.Fatalf("expected student ID update to succeed, got %v", err)
	}

	if len(store.studentInfoUpdates) != 1 {
		t.Fatalf("expected one student info update, got %#v", store.studentInfoUpdates)
	}
	if update := store.studentInfoUpdates[0]; !update.isStudent || update.studentID != "87654321" {
		t.Fatalf("unexpected student info update: %#v", update)
	}
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeSuccess, actionStudentIDUpdated)
}

func TestUpdateUserRejectsInvalidStudentIDFormat(t *testing.T) {
	store := newFakeAdminStore(t, true, "12345678", db.RoleTypeMember, "member")

	err := updateUser(t, store, UpdateUserRequest{StudentID: ptr("1234")})

	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(store.studentInfoUpdates) != 0 {
		t.Fatalf("expected no student info update, got %#v", store.studentInfoUpdates)
	}
}

/*
	Student status
*/

func TestUpdateUserRejectsBecomingStudentWithoutStudentID(t *testing.T) {
	store := newFakeAdminStore(t, false, "N1234567", db.RoleTypeMember, "member")

	err := updateUser(t, store, UpdateUserRequest{IsStudent: ptr(true)})

	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(store.studentInfoUpdates) != 0 {
		t.Fatalf("expected no student info update, got %#v", store.studentInfoUpdates)
	}
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeFailed, actionStudentStatusUpdated)
}

func TestUpdateUserRejectsBecomingStudentWithNonEightDigitID(t *testing.T) {
	store := newFakeAdminStore(t, false, "N1234567", db.RoleTypeMember, "member")

	err := updateUser(t, store, UpdateUserRequest{IsStudent: ptr(true), StudentID: ptr("1234567")})

	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(store.studentInfoUpdates) != 0 {
		t.Fatalf("expected no student info update, got %#v", store.studentInfoUpdates)
	}
}

func TestUpdateUserBecomesStudentWithValidStudentID(t *testing.T) {
	store := newFakeAdminStore(t, false, "N1234567", db.RoleTypeMember, "member")

	if err := updateUser(t, store, UpdateUserRequest{IsStudent: ptr(true), StudentID: ptr("12345678")}); err != nil {
		t.Fatalf("expected student status update to succeed, got %v", err)
	}

	if len(store.studentInfoUpdates) != 1 {
		t.Fatalf("expected one student info update, got %#v", store.studentInfoUpdates)
	}
	if update := store.studentInfoUpdates[0]; !update.isStudent || update.studentID != "12345678" {
		t.Fatalf("unexpected student info update: %#v", update)
	}
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeSuccess, actionStudentStatusUpdated)
}

func TestUpdateUserDroppingStudentStatusGeneratesNonStudentID(t *testing.T) {
	store := newFakeAdminStore(t, true, "12345678", db.RoleTypeMember, "member")

	if err := updateUser(t, store, UpdateUserRequest{IsStudent: ptr(false)}); err != nil {
		t.Fatalf("expected student status update to succeed, got %v", err)
	}

	if len(store.studentInfoUpdates) != 1 {
		t.Fatalf("expected one student info update, got %#v", store.studentInfoUpdates)
	}

	update := store.studentInfoUpdates[0]
	if update.isStudent {
		t.Fatal("expected user to no longer be a student")
	}
	if !nonStudentIDRegex.MatchString(update.studentID) {
		t.Fatalf("expected a generated non-student ID, got %q", update.studentID)
	}
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeSuccess, actionStudentStatusUpdated)
}

func TestUpdateUserFailsWhenNoUnusedNonStudentIDIsAvailable(t *testing.T) {
	store := newFakeAdminStore(t, true, "12345678", db.RoleTypeMember, "member")

	// Every ID the generator can produce is taken, so generation gives up
	// rather than writing a duplicate.
	store.allStudentIDsTaken = true

	err := updateUser(t, store, UpdateUserRequest{IsStudent: ptr(false)})
	if err == nil {
		t.Fatal("expected non-student ID generation to fail")
	}
	if len(store.studentInfoUpdates) != 0 {
		t.Fatalf("expected no student info update, got %#v", store.studentInfoUpdates)
	}
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeFailed, actionStudentStatusUpdated)
}

/*
	Groups
*/

func TestUpdateUserIgnoresMemberGroupRemoval(t *testing.T) {
	store := newFakeAdminStore(t, true, "12345678", db.RoleTypeMember, "member", "board")

	if err := updateUser(t, store, UpdateUserRequest{GroupsRemove: []db.GroupType{db.GroupTypeMember}}); err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}

	if len(store.removedGroups) != 0 {
		t.Fatalf("expected member group removal to be ignored, got %v", store.removedGroups)
	}
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeSuccess)
}

func TestUpdateUserRemovesNonMemberGroup(t *testing.T) {
	store := newFakeAdminStore(t, true, "12345678", db.RoleTypeMember, "member", "board")

	if err := updateUser(t, store, UpdateUserRequest{
		GroupsRemove: []db.GroupType{db.GroupTypeMember, db.GroupTypeBoard},
	}); err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}

	if len(store.removedGroups) != 1 || store.removedGroups[0] != db.GroupTypeBoard {
		t.Fatalf("expected only the board group to be removed, got %v", store.removedGroups)
	}
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeSuccess, actionGroupRemoved)
}

func TestUpdateUserAddsGroups(t *testing.T) {
	store := newFakeAdminStore(t, true, "12345678", db.RoleTypeMember, "member")

	if err := updateUser(t, store, UpdateUserRequest{
		GroupsAdd: []db.GroupType{db.GroupTypeMember, db.GroupTypeExecutive},
	}); err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}

	// The member group is already assigned, so only executive is added.
	if len(store.addedGroups) != 1 || store.addedGroups[0] != db.GroupTypeExecutive {
		t.Fatalf("expected only the executive group to be added, got %v", store.addedGroups)
	}
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeSuccess, actionGroupAdded)
}

func TestUpdateUserRejectsUnknownGroup(t *testing.T) {
	store := newFakeAdminStore(t, true, "12345678", db.RoleTypeMember, "member")

	err := updateUser(t, store, UpdateUserRequest{GroupsAdd: []db.GroupType{db.GroupType("wizards")}})

	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(store.addedGroups) != 0 {
		t.Fatalf("expected no group to be added, got %v", store.addedGroups)
	}
}

/*
	Role
*/

func TestUpdateUserFloorsClearedRoleToMember(t *testing.T) {
	store := newFakeAdminStore(t, true, "12345678", db.RoleTypeAdmin, "member")

	if err := updateUser(t, store, UpdateUserRequest{Role: ptr(db.RoleType(""))}); err != nil {
		t.Fatalf("expected role update to succeed, got %v", err)
	}

	if len(store.roleUpdates) != 1 || store.roleUpdates[0] != db.RoleTypeMember {
		t.Fatalf("expected role to be floored to member, got %v", store.roleUpdates)
	}
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeSuccess, actionRoleUpdated)
}

func TestUpdateUserSetsAdminRole(t *testing.T) {
	store := newFakeAdminStore(t, true, "12345678", db.RoleTypeMember, "member")

	if err := updateUser(t, store, UpdateUserRequest{Role: ptr(db.RoleTypeAdmin)}); err != nil {
		t.Fatalf("expected role update to succeed, got %v", err)
	}

	if len(store.roleUpdates) != 1 || store.roleUpdates[0] != db.RoleTypeAdmin {
		t.Fatalf("expected role to be set to admin, got %v", store.roleUpdates)
	}
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeSuccess, actionRoleUpdated)
}

func TestUpdateUserRejectsInvalidRole(t *testing.T) {
	store := newFakeAdminStore(t, true, "12345678", db.RoleTypeMember, "member")

	err := updateUser(t, store, UpdateUserRequest{Role: ptr(db.RoleType("superadmin"))})

	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(store.roleUpdates) != 0 {
		t.Fatalf("expected no role update, got %v", store.roleUpdates)
	}
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeFailed, actionRoleUpdated)
}

func TestUpdateUserLeavesRoleUntouchedWhenAbsent(t *testing.T) {
	store := newFakeAdminStore(t, true, "12345678", db.RoleTypeAdmin, "member")

	if err := updateUser(t, store, UpdateUserRequest{GroupsAdd: []db.GroupType{db.GroupTypeBoard}}); err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}

	if len(store.roleUpdates) != 0 {
		t.Fatalf("expected role to be left alone, got %v", store.roleUpdates)
	}
}

/*
	Membership
*/

func TestUpdateUserCancelMembershipWithoutActiveMembership(t *testing.T) {
	store := newFakeAdminStore(t, true, "12345678", db.RoleTypeMember, "member")

	err := updateUser(t, store, UpdateUserRequest{CancelMembershipId: ptr(testMembershipID)})

	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(store.cancelledMembershipIDs) != 1 || store.cancelledMembershipIDs[0] != testMembershipID {
		t.Fatalf("expected cancellation attempt for %q, got %v", testMembershipID, store.cancelledMembershipIDs)
	}
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeFailed, actionMembershipCancelled)
}

func TestUpdateUserCancelsMembership(t *testing.T) {
	store := newFakeAdminStore(t, true, "12345678", db.RoleTypeMember, "member")
	store.cancelMembershipResult = true

	if err := updateUser(t, store, UpdateUserRequest{CancelMembershipId: ptr(testMembershipID)}); err != nil {
		t.Fatalf("expected cancellation to succeed, got %v", err)
	}

	if len(store.cancelledMembershipUsers) != 1 || store.cancelledMembershipUsers[0] != testTargetID {
		t.Fatalf("expected cancellation for user %q, got %v", testTargetID, store.cancelledMembershipUsers)
	}
	if len(store.cancelledMembershipIDs) != 1 || store.cancelledMembershipIDs[0] != testMembershipID {
		t.Fatalf("expected cancellation for membership %q, got %v", testMembershipID, store.cancelledMembershipIDs)
	}
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeSuccess, actionMembershipCancelled)
}

func TestUpdateUserRejectsBlankMembershipID(t *testing.T) {
	store := newFakeAdminStore(t, true, "12345678", db.RoleTypeMember, "member")

	err := updateUser(t, store, UpdateUserRequest{CancelMembershipId: ptr("  ")})

	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(store.cancelledMembershipIDs) != 0 {
		t.Fatalf("expected no cancellation attempt, got %v", store.cancelledMembershipIDs)
	}
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeFailed, actionMembershipCancelled)
}

func TestUpdateUserRejectsInvalidMembershipID(t *testing.T) {
	store := newFakeAdminStore(t, true, "12345678", db.RoleTypeMember, "member")

	err := updateUser(t, store, UpdateUserRequest{CancelMembershipId: ptr("not-a-uuid")})

	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(store.cancelledMembershipIDs) != 0 {
		t.Fatalf("expected no cancellation attempt, got %v", store.cancelledMembershipIDs)
	}
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeFailed, actionMembershipCancelled)
}

/*
	Audit logging
*/

func TestUpdateUserLogsOneEntryPerAction(t *testing.T) {
	store := newFakeAdminStore(t, true, "12345678", db.RoleTypeMember, "member")

	if err := updateUser(t, store, UpdateUserRequest{
		Role:      ptr(db.RoleTypeAdmin),
		GroupsAdd: []db.GroupType{db.GroupTypeBoard},
	}); err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}

	assertAuditActions(t, store, db.AdminAuditOutcomeTypeSuccess, actionRoleUpdated, actionGroupAdded)

	for _, log := range store.auditLogs {
		if !log.TargetUserID.Valid {
			t.Fatalf("expected audit entry %q to name the target user", log.Action)
		}
		if log.RequestID != testRequestID {
			t.Fatalf("expected audit entry %q to carry the request ID, got %q", log.Action, log.RequestID)
		}
		if !log.Description.Valid || log.Description.String == "" {
			t.Fatalf("expected audit entry %q to carry a description", log.Action)
		}
	}
}

func TestUpdateUserRollsBackAndLogsFailureWhenLaterActionFails(t *testing.T) {
	store := newFakeAdminStore(t, true, "12345678", db.RoleTypeMember, "member")

	err := updateUser(t, store, UpdateUserRequest{
		Role:               ptr(db.RoleTypeAdmin),
		CancelMembershipId: ptr(testMembershipID),
	})

	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}

	// The role change was rolled back with the transaction, so the only entry
	// left is the failure of the action that broke the request.
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeSuccess)
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeFailed, actionMembershipCancelled)
}

func TestUpdateUserReturnsNotFoundForUnknownUser(t *testing.T) {
	store := newFakeAdminStore(t, true, "12345678", db.RoleTypeMember, "member")
	store.getUserErr = pgx.ErrNoRows

	err := updateUser(t, store, UpdateUserRequest{Role: ptr(db.RoleTypeAdmin)})

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
	assertAuditActions(t, store, db.AdminAuditOutcomeTypeFailed, actionUserUpdated)
}
