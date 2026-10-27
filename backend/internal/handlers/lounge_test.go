package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestUpdateStationRejectsInvalidUUID(t *testing.T) {
	handler := &LoungeHandler{}

	r := chi.NewRouter()
	r.Patch("/lounge/stations/{id}", handler.UpdateStation)

	req := httptest.NewRequest(http.MethodPatch, "/lounge/stations/not-a-valid-uuid", bytes.NewBufferString(`{"closed":true}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUpdateStationRejectsMalformedBody(t *testing.T) {
	handler := &LoungeHandler{}

	r := chi.NewRouter()
	r.Patch("/lounge/stations/{id}", handler.UpdateStation)

	req := httptest.NewRequest(http.MethodPatch, "/lounge/stations/c1000000-0000-4000-8000-000000000001", bytes.NewBufferString(`invalid json`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}
