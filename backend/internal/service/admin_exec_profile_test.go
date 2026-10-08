package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ubcesports/memberships/internal/database/db"
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
