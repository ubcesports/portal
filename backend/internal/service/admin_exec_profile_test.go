package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/dto"
)

func TestUpdateExecProfileAcceptsEveryDisplayGroup(t *testing.T) {
	displayGroups := []db.ExecDisplayGroupType{
		db.ExecDisplayGroupTypePresident,
		db.ExecDisplayGroupTypeBoard,
		db.ExecDisplayGroupTypeCentralDirector,
		db.ExecDisplayGroupTypeGameDirector,
		db.ExecDisplayGroupTypeExecutive,
	}

	for _, displayGroup := range displayGroups {
		t.Run(string(displayGroup), func(t *testing.T) {
			store := newFakeAdminStore(t, false, "N1234567", db.RoleTypeMember, "executive")
			adminService := &AdminService{adminRepository: store}

			_, err := adminService.UpdateExecProfile(
				context.Background(),
				testActorID,
				testTargetID,
				pgtype.Text{},
				pgtype.Int4{},
				db.NullExecDisplayGroupType{ExecDisplayGroupType: displayGroup, Valid: true},
				testRequestID,
			)
			if err != nil {
				t.Fatalf("expected display group %q to be accepted, got %v", displayGroup, err)
			}
			if len(store.execProfileUpdates) != 1 {
				t.Fatalf("expected one exec profile update, got %d", len(store.execProfileUpdates))
			}
			if got := store.execProfileUpdates[0].displayGroup; !got.Valid || got.ExecDisplayGroupType != displayGroup {
				t.Fatalf("expected display group %q, got %#v", displayGroup, got)
			}
		})
	}
}

func TestGetExecProfileReturnsAdminDTO(t *testing.T) {
	store := newFakeAdminStore(t, false, "N1234567", db.RoleTypeMember, "executive")
	store.execProfile.Title = "VP Events"
	store.execProfile.DisplayOrder = 4
	store.execProfile.DisplayGroup = db.ExecDisplayGroupTypeCentralDirector
	adminService := &AdminService{adminRepository: store}

	profile, err := adminService.GetExecProfile(context.Background(), testTargetID)
	if err != nil {
		t.Fatalf("expected profile, got %v", err)
	}
	if profile.Title != "VP Events" || profile.DisplayOrder != 4 || profile.DisplayGroup != dto.ExecDisplayGroupTypeCentralDirector {
		t.Fatalf("unexpected profile: %#v", profile)
	}
}

func TestGetExecProfileMapsMissingProfileToNotFound(t *testing.T) {
	store := newFakeAdminStore(t, false, "N1234567", db.RoleTypeMember, "member")
	store.getExecProfileErr = pgx.ErrNoRows
	adminService := &AdminService{adminRepository: store}

	_, err := adminService.GetExecProfile(context.Background(), testTargetID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestUpdateExecProfileAllowsOmittedDisplayGroup(t *testing.T) {
	store := newFakeAdminStore(t, false, "N1234567", db.RoleTypeMember, "executive")
	adminService := &AdminService{adminRepository: store}

	_, err := adminService.UpdateExecProfile(
		context.Background(),
		testActorID,
		testTargetID,
		pgtype.Text{String: "President", Valid: true},
		pgtype.Int4{Int32: 1, Valid: true},
		db.NullExecDisplayGroupType{},
		testRequestID,
	)
	if err != nil {
		t.Fatalf("expected omitted display group to be accepted, got %v", err)
	}
	if len(store.execProfileUpdates) != 1 || store.execProfileUpdates[0].displayGroup.Valid {
		t.Fatalf("expected a null display group update, got %#v", store.execProfileUpdates)
	}
}

func TestUpdateExecProfileRejectsInvalidDisplayGroup(t *testing.T) {
	store := newFakeAdminStore(t, false, "N1234567", db.RoleTypeMember, "executive")
	adminService := &AdminService{adminRepository: store}

	_, err := adminService.UpdateExecProfile(
		context.Background(),
		testActorID,
		testTargetID,
		pgtype.Text{},
		pgtype.Int4{},
		db.NullExecDisplayGroupType{ExecDisplayGroupType: "invalid", Valid: true},
		testRequestID,
	)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(store.execProfileUpdates) != 0 {
		t.Fatalf("expected no exec profile update, got %#v", store.execProfileUpdates)
	}
}

func TestUpdateExecProfileRejectsNegativeDisplayOrder(t *testing.T) {
	store := newFakeAdminStore(t, false, "N1234567", db.RoleTypeMember, "executive")
	adminService := &AdminService{adminRepository: store}

	_, err := adminService.UpdateExecProfile(
		context.Background(),
		testActorID,
		testTargetID,
		pgtype.Text{},
		pgtype.Int4{Int32: -1, Valid: true},
		db.NullExecDisplayGroupType{},
		testRequestID,
	)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(store.execProfileUpdates) != 0 {
		t.Fatalf("expected no exec profile update, got %#v", store.execProfileUpdates)
	}
}
