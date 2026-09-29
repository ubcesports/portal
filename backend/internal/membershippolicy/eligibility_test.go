package membershippolicy

import (
	"testing"
	"time"

	"github.com/ubcesports/memberships/internal/dto"
)

func vancouverTime(t *testing.T, year int, month time.Month, day, hour int) time.Time {
	t.Helper()
	loc, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	return time.Date(year, month, day, hour, 0, 0, 0, loc)
}

func TestIsPurchaseClosedYear(t *testing.T) {
	tests := []struct {
		name string
		now  time.Time
		want bool
	}{
		{"well before closure", vancouverTime(t, 2026, time.March, 1, 12), false},
		{"first closed day", vancouverTime(t, 2026, time.April, 28, 0), true},
		{"last closed day", vancouverTime(t, 2026, time.April, 30, 23), true},
		{"reopens May 1", vancouverTime(t, 2026, time.May, 1, 0), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsPurchaseClosed(tt.now, dto.MembershipExpirationYear)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("IsPurchaseClosed = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPurchaseClosedSemester(t *testing.T) {
	tests := []struct {
		name string
		now  time.Time
		want bool
	}{
		{"mid fall semester", vancouverTime(t, 2026, time.October, 15, 12), false},
		{"winter closure starts Dec 30", vancouverTime(t, 2026, time.December, 30, 0), true},
		{"winter closure last day", vancouverTime(t, 2026, time.December, 31, 23), true},
		{"reopens Jan 1", vancouverTime(t, 2027, time.January, 1, 0), false},
		{"summer closure starts Apr 28", vancouverTime(t, 2026, time.April, 28, 0), true},
		{"summer closure mid", vancouverTime(t, 2026, time.July, 1, 12), true},
		{"reopens Sep 1", vancouverTime(t, 2026, time.September, 1, 0), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsPurchaseClosed(tt.now, dto.MembershipExpirationSemester)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("IsPurchaseClosed = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPurchaseClosedDay(t *testing.T) {
	tests := []struct {
		name string
		now  time.Time
		want bool
	}{
		{"open in fall", vancouverTime(t, 2026, time.October, 15, 12), false},
		{"closed starting May 1", vancouverTime(t, 2026, time.May, 1, 0), true},
		{"closed mid summer", vancouverTime(t, 2026, time.July, 4, 12), true},
		{"reopens Sep 1", vancouverTime(t, 2026, time.September, 1, 0), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsPurchaseClosed(tt.now, dto.MembershipExpirationDay)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("IsPurchaseClosed = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNextPurchaseOpenDate(t *testing.T) {
	tests := []struct {
		name           string
		now            time.Time
		expirationType dto.MembershipExpirationType
		wantMonth      time.Month
		wantDay        int
		wantYear       int
	}{
		{"year closure reopens same year", vancouverTime(t, 2026, time.April, 29, 12), dto.MembershipExpirationYear, time.May, 1, 2026},
		{"semester winter closure reopens next January", vancouverTime(t, 2026, time.December, 30, 12), dto.MembershipExpirationSemester, time.January, 1, 2027},
		{"semester summer closure reopens September", vancouverTime(t, 2026, time.June, 1, 12), dto.MembershipExpirationSemester, time.September, 1, 2026},
		{"day closure reopens September", vancouverTime(t, 2026, time.July, 4, 12), dto.MembershipExpirationDay, time.September, 1, 2026},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NextPurchaseOpenDate(tt.now, tt.expirationType)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Month() != tt.wantMonth || got.Day() != tt.wantDay || got.Year() != tt.wantYear {
				t.Fatalf("got %s, want %s %d, %d", got.Format(time.DateOnly), tt.wantMonth, tt.wantDay, tt.wantYear)
			}
		})
	}
}

func TestIsUserGroupEligible(t *testing.T) {
	tests := []struct {
		name          string
		userGroups    []dto.GroupType
		requiredGroup dto.GroupType
		want          bool
	}{
		{"member matches member requirement", []dto.GroupType{dto.GroupMember}, dto.GroupMember, true},
		{"member does not match executive requirement", []dto.GroupType{dto.GroupMember}, dto.GroupExecutive, false},
		{"director satisfies executive requirement", []dto.GroupType{dto.GroupMember, dto.GroupDirector}, dto.GroupExecutive, true},
		{"board satisfies executive requirement", []dto.GroupType{dto.GroupMember, dto.GroupBoard}, dto.GroupExecutive, true},
		{"competitive team does not satisfy executive requirement", []dto.GroupType{dto.GroupCompetitiveTeam}, dto.GroupExecutive, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUserGroupEligible(tt.userGroups, tt.requiredGroup)
			if got != tt.want {
				t.Fatalf("isUserGroupEligible = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindCurrentMembershipForProgram(t *testing.T) {
	memberships := []dto.MembershipDTO{
		{ID: "m1", ProgramId: "general", Slug: "basic"},
		{ID: "m2", ProgramId: "ssbm", Slug: "ssbm_yearly"},
	}

	found, err := findCurrentMembershipForProgram(memberships, "general")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil || found.ID != "m1" {
		t.Fatalf("expected to find m1, got %+v", found)
	}

	notFound, err := findCurrentMembershipForProgram(memberships, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notFound != nil {
		t.Fatalf("expected no match, got %+v", notFound)
	}
}

func TestFindCurrentMembershipForProgramRejectsMultipleActive(t *testing.T) {
	memberships := []dto.MembershipDTO{
		{ID: "m1", ProgramId: "general", Slug: "basic"},
		{ID: "m2", ProgramId: "general", Slug: "lounge"},
	}

	_, err := findCurrentMembershipForProgram(memberships, "general")
	if err == nil {
		t.Fatal("expected an error when a user has two active memberships in the same program")
	}
}
