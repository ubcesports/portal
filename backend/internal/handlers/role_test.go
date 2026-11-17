package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ubcesports/memberships/internal/service"
)

func TestRoleWriteErrorMapsServiceErrors(t *testing.T) {
	handler := &RoleHandler{}

	for _, tc := range []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"invalid role", service.ErrInvalidRole, http.StatusBadRequest, "INVALID_ROLE"},
		{"invalid permission", service.ErrInvalidRolePermission, http.StatusBadRequest, "INVALID_ROLE_PERMISSION"},
		{"name taken", service.ErrRoleNameTaken, http.StatusConflict, "ROLE_NAME_TAKEN"},
		{"role not found", service.ErrRoleNotFound, http.StatusNotFound, "ROLE_NOT_FOUND"},
		{"user not found", service.ErrRoleUserNotFound, http.StatusNotFound, "USER_NOT_FOUND"},
		{"already assigned", service.ErrRoleAlreadyAssigned, http.StatusConflict, "ROLE_ALREADY_ASSIGNED"},
		{"not assigned", service.ErrRoleNotAssigned, http.StatusNotFound, "ROLE_NOT_ASSIGNED"},
		{"wrapped", errors.Join(errors.New("context"), service.ErrRoleNotAssigned), http.StatusNotFound, "ROLE_NOT_ASSIGNED"},
		{"unexpected", errors.New("database failure"), http.StatusInternalServerError, "INTERNAL_ERROR"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.writeError(rec, httptest.NewRequest(http.MethodPost, "/", nil), tc.err, "assign")

			if rec.Code != tc.status {
				t.Fatalf("expected status %d, got %d", tc.status, rec.Code)
			}
			var body struct {
				Code string `json:"code"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Code != tc.code {
				t.Fatalf("expected code %s, got %s", tc.code, body.Code)
			}
		})
	}
}

func TestAssignAndUnassignRequireAuthentication(t *testing.T) {
	handler := &RoleHandler{}

	r := chi.NewRouter()
	r.Post("/admin/roles/{id}/users", handler.AssignUser)
	r.Delete("/admin/roles/{id}/users/{userId}", handler.UnassignUser)

	for _, tc := range []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"assign", http.MethodPost, "/admin/roles/11111111-2222-4333-8444-555555555555/users", `{"user_id":"3c2b1a09-8765-4321-abcd-0123456789ab"}`},
		{"unassign", http.MethodDelete, "/admin/roles/11111111-2222-4333-8444-555555555555/users/3c2b1a09-8765-4321-abcd-0123456789ab", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body)))

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
			}
		})
	}
}
