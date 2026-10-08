package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExecProfileHandlerRejectsUnauthorizedCurrentProfileRequests(t *testing.T) {
	handler := &ExecProfileHandler{}
	req := httptest.NewRequest(http.MethodGet, "/exec-profile", nil)
	rec := httptest.NewRecorder()

	handler.GetCurrentExecProfile(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}
