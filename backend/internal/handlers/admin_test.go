package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/dto"
)

func TestParseAdminFiltersForExportIgnoresPagination(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/users/export?full_name=dip&role=member&group=competitive_team&group=member&is_student=false&limit=invalid",
		nil,
	)

	filters, err := parseAdminUserFilters(req, false)
	if err != nil {
		t.Fatalf("expected valid filters, got %v", err)
	}
	if filters.FullName != "dip" || filters.Role != "member" || len(filters.Groups) != 2 || filters.Groups[0] != "competitive_team" || filters.Groups[1] != "member" {
		t.Fatalf("unexpected filters: %#v", filters)
	}
	if filters.IsStudent == nil || *filters.IsStudent {
		t.Fatalf("expected is_student=false, got %#v", filters.IsStudent)
	}
	if filters.Limit != 0 || filters.Offset != 0 {
		t.Fatalf("expected export pagination to be disabled, got limit=%d offset=%d", filters.Limit, filters.Offset)
	}
}

func TestParseAdminFiltersAcceptsRepeatedMembershipTiers(t *testing.T) {
	first := "2d746a56-c977-49e0-a04c-20504cdb07c0"
	second := "3c2b1a09-8765-4321-abcd-0123456789ab"
	req := httptest.NewRequest(http.MethodGet, "/admin/users?membership_tier_id="+first+"&membership_tier_id="+second, nil)

	filters, err := parseAdminUserFilters(req, true)
	if err != nil {
		t.Fatalf("expected valid filters, got %v", err)
	}
	if len(filters.MembershipTierIDs) != 2 || filters.MembershipTierIDs[0] != first || filters.MembershipTierIDs[1] != second {
		t.Fatalf("unexpected tier filters: %#v", filters.MembershipTierIDs)
	}
}

func TestParseAdminFiltersRejectsInvalidMembershipTier(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/users?membership_tier_id=invalid", nil)

	if _, err := parseAdminUserFilters(req, true); err == nil {
		t.Fatal("expected invalid membership tier error")
	}
}

func TestParseAdminFiltersRejectsInvalidGroup(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/users?group=invalid", nil)

	if _, err := parseAdminUserFilters(req, true); err == nil {
		t.Fatal("expected invalid group error")
	}
}

func TestParseAdminAuditLogFilters(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/audit-logs?actor_name=dip&limit=10&offset=20", nil)

	filters, err := parseAdminAuditLogFilters(req)
	if err != nil {
		t.Fatalf("expected valid filters, got %v", err)
	}
	if filters.ActorName != "dip" || filters.Limit != 10 || filters.Offset != 20 {
		t.Fatalf("unexpected filters: %#v", filters)
	}
}

func TestParseAdminAuditLogFiltersRejectsInvalidPagination(t *testing.T) {
	tests := []string{
		"/admin/audit-logs?limit=0",
		"/admin/audit-logs?offset=-1",
	}

	for _, target := range tests {
		req := httptest.NewRequest(http.MethodGet, target, nil)
		if _, err := parseAdminAuditLogFilters(req); err == nil {
			t.Fatalf("expected invalid pagination error for %s", target)
		}
	}
}

func TestBuildUpdateUserRequestMapsBodyToServiceTypes(t *testing.T) {
	studentID := "12345678"
	adminNote := "Follow up about membership"
	membershipID := "2d746a56-c977-49e0-a04c-20504cdb07c0"
	isStudent := true
	role := dto.RoleAdmin

	request := buildUpdateUserRequest(dto.AdminUpdateUserRequest{
		StudentID:          &studentID,
		IsStudent:          &isStudent,
		GroupsAdd:          []dto.GroupType{dto.GroupBoard},
		GroupsRemove:       []dto.GroupType{dto.GroupMember, dto.GroupExecutive},
		Role:               &role,
		CancelMembershipId: &membershipID,
		AdminNote:          &adminNote,
	})

	if request.StudentID == nil || *request.StudentID != "12345678" {
		t.Fatalf("expected student ID to be carried over, got %#v", request.StudentID)
	}
	if request.IsStudent == nil || !*request.IsStudent {
		t.Fatalf("expected is_student=true, got %#v", request.IsStudent)
	}
	if request.Role == nil || *request.Role != db.RoleTypeAdmin {
		t.Fatalf("expected admin role, got %#v", request.Role)
	}
	if len(request.GroupsAdd) != 1 || request.GroupsAdd[0] != db.GroupTypeBoard {
		t.Fatalf("unexpected groups to add: %v", request.GroupsAdd)
	}
	if len(request.GroupsRemove) != 2 || request.GroupsRemove[1] != db.GroupTypeExecutive {
		t.Fatalf("unexpected groups to remove: %v", request.GroupsRemove)
	}
	if request.CancelMembershipId == nil || *request.CancelMembershipId != membershipID {
		t.Fatalf("expected membership ID to be carried over, got %#v", request.CancelMembershipId)
	}
	if request.AdminNote == nil || *request.AdminNote != adminNote {
		t.Fatalf("expected admin note to be carried over, got %#v", request.AdminNote)
	}
}

func TestBuildUpdateUserRequestLeavesAbsentRoleNil(t *testing.T) {
	request := buildUpdateUserRequest(dto.AdminUpdateUserRequest{})

	if request.Role != nil {
		t.Fatalf("expected an absent role to stay nil, got %#v", request.Role)
	}
}

func TestSafeCSVCellEscapesFormulaPrefix(t *testing.T) {
	if got := safeCSVCell("=1+1"); got != "'=1+1" {
		t.Fatalf("expected formula prefix to be escaped, got %q", got)
	}
	if got := safeCSVCell("Sudipto Islam"); got != "Sudipto Islam" {
		t.Fatalf("expected ordinary text to be unchanged, got %q", got)
	}
}
