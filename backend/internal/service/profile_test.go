package service

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/repository"
)

func TestOnboardUserSavesProvidedName(t *testing.T) {
	for _, isStudent := range []bool{false, true} {
		t.Run(map[bool]string{false: "community member", true: "student"}[isStudent], func(t *testing.T) {
			store := &onboardingTestDB{profile: db.GetProfileByUserIDRow{FullName: "Provider Name"}}
			service := NewProfileService(repository.NewProfileRepository(db.New(store)))
			studentID := "12345678"
			err := service.OnboardUser(context.Background(), testTargetID, dto.OnboardUserRequest{
				FullName: "  Renée O'Connor-Smith  ", IsStudent: isStudent, StudentID: &studentID,
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(store.onboardArgs) != 5 || store.onboardArgs[3] != "Renée O'Connor-Smith" {
				t.Fatalf("expected trimmed user name in onboarding query, got %#v", store.onboardArgs)
			}
			if store.onboardArgs[1] != isStudent {
				t.Fatalf("student status was not preserved: %#v", store.onboardArgs)
			}
		})
	}
}

func TestOnboardUserPassesExecutiveInviteToAtomicOnboardingQuery(t *testing.T) {
	t.Setenv("EXEC_ONBOARDING_CODE", "exec-invite")
	store := &onboardingTestDB{profile: db.GetProfileByUserIDRow{FullName: "Provider Name"}}
	service := NewProfileService(repository.NewProfileRepository(db.New(store)))
	inviteCode := " exec-invite "

	err := service.OnboardUser(context.Background(), testTargetID, dto.OnboardUserRequest{
		FullName:   "Executive User",
		InviteCode: &inviteCode,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(store.onboardArgs) != 5 || store.onboardArgs[0] != true {
		t.Fatalf("expected executive onboarding query, got %#v", store.onboardArgs)
	}
}

func TestOnboardUserCannotChangeNameAfterOnboarding(t *testing.T) {
	store := &onboardingTestDB{profile: db.GetProfileByUserIDRow{
		FullName:              "Original Name",
		OnboardingCompletedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}}
	service := NewProfileService(repository.NewProfileRepository(db.New(store)))
	err := service.OnboardUser(context.Background(), testTargetID, dto.OnboardUserRequest{FullName: "Replacement Name"})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected conflict for completed onboarding, got %v", err)
	}
	if store.onboardArgs != nil {
		t.Fatal("completed onboarding must not write a replacement name")
	}
}

// Exercise the service, repository, and generated argument bindings without a database.
type onboardingTestDB struct {
	db.DBTX
	profile     db.GetProfileByUserIDRow
	onboardArgs []any
}

func (s *onboardingTestDB) Exec(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	if strings.HasPrefix(query, "-- name: OnboardUserByUserId ") {
		s.onboardArgs = args
	}
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}

func (s *onboardingTestDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return onboardingTestRow{profile: s.profile}
}

type onboardingTestRow struct {
	profile db.GetProfileByUserIDRow
}

func (r onboardingTestRow) Scan(dest ...any) error {
	values := reflect.ValueOf(r.profile)
	for i, target := range dest {
		reflect.ValueOf(target).Elem().Set(values.Field(i))
	}
	return nil
}

func TestOnboardUserRequiresFullName(t *testing.T) {
	for _, name := range []string{"", "   ", "\t\n", "\u00a0"} {
		t.Run(name, func(t *testing.T) {
			service := NewProfileService(nil)
			err := service.OnboardUser(context.Background(), "", dto.OnboardUserRequest{FullName: name})
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("expected validation error for blank full name, got %v", err)
			}
		})
	}
}

func TestProfileServiceGetProfileByUserIDAggregatesGroups(t *testing.T) {
	now := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	userID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	studentID := pgtype.Text{String: "12345678", Valid: true}
	emailVerifiedAt := pgtype.Timestamptz{Time: now.Add(-24 * time.Hour), Valid: true}
	onboardingCompletedAt := pgtype.Timestamptz{Time: now.Add(-48 * time.Hour), Valid: true}
	avatarURL := pgtype.Text{String: "https://example.com/avatar.png", Valid: true}

	row := db.GetProfileByUserIDRow{
		ID:                    userID,
		Email:                 "sudi@example.com",
		StudentID:             studentID,
		Role:                  db.RoleTypeMember,
		CreatedAt:             pgtype.Timestamptz{Time: now.Add(-72 * time.Hour), Valid: true},
		UpdatedAt:             pgtype.Timestamptz{Time: now, Valid: true},
		FullName:              "Sudi Mango",
		EmailVerifiedAt:       emailVerifiedAt,
		IsStudent:             true,
		OnboardingCompletedAt: onboardingCompletedAt,
		AvatarUrl:             avatarURL,
		Groups:                []string{"member", "board"},
	}

	profile := buildProfile(row)

	if len(profile.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(profile.Groups))
	}
	if profile.Groups[0] != "member" || profile.Groups[1] != "board" {
		t.Fatalf("expected ordered groups [member board], got %v", profile.Groups)
	}
	if profile.StudentID == nil || *profile.StudentID != "12345678" {
		t.Fatalf("expected student ID to be preserved, got %#v", profile.StudentID)
	}
}

func TestProfileServiceGetProfileByUserIDReturnsEmptyGroupsSlice(t *testing.T) {
	now := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	userID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}

	profile := buildProfile(db.GetProfileByUserIDRow{
		ID:        userID,
		Email:     "sudi@example.com",
		Role:      db.RoleTypeMember,
		CreatedAt: pgtype.Timestamptz{Time: now.Add(-72 * time.Hour), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		FullName:  "Sudi Mango",
		IsStudent: false,
		Groups:    []string{},
	})

	if profile.Groups == nil {
		t.Fatal("expected groups to be an empty slice, got nil")
	}
	if len(profile.Groups) != 0 {
		t.Fatalf("expected 0 groups, got %d", len(profile.Groups))
	}
}
