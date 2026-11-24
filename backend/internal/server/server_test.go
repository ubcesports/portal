package server

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
)

func TestRequestLoggerAddsRequestIDAndLogsRequest(t *testing.T) {
	var logs bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	handler := middleware.RequestID(logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})))
	req := httptest.NewRequest(http.MethodPost, "/onboard", nil)
	req.RemoteAddr = "203.0.113.7:54321"
	req.Header.Set("X-Forwarded-For", "198.51.100.1, 10.0.0.2")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected an X-Request-ID response header")
	}
	for _, expected := range []string{
		`"msg":"request completed"`,
		`"path":"/onboard"`,
		`"status":201`,
		`"remote_ip":"203.0.113.7"`,
		`"forwarded_for":"198.51.100.1, 10.0.0.2"`,
	} {
		if !strings.Contains(logs.String(), expected) {
			t.Fatalf("expected log to contain %s, got %s", expected, logs.String())
		}
	}
}

func TestRecovererReturns500AndLogsPanic(t *testing.T) {
	var logs bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	handler := middleware.RequestID(logger(recoverer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("test panic")
	}))))
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected an X-Request-ID response header")
	}
	for _, expected := range []string{`"msg":"panic recovered"`, `"panic":"test panic"`, `"stack":`, `"remote_ip":"192.0.2.1"`} {
		if !strings.Contains(logs.String(), expected) {
			t.Fatalf("expected panic log to contain %s, got %s", expected, logs.String())
		}
	}
}

func TestRequestLoggerOmitsForwardedForWhenAbsent(t *testing.T) {
	var logs bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	handler := middleware.RequestID(logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req.RemoteAddr = "[2001:db8::1]:443"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !strings.Contains(logs.String(), `"remote_ip":"2001:db8::1"`) {
		t.Fatalf("expected log to contain IPv6 remote_ip, got %s", logs.String())
	}
	if strings.Contains(logs.String(), `"forwarded_for"`) {
		t.Fatalf("expected no forwarded_for attr, got %s", logs.String())
	}
}

func TestCORSMiddlewareAllowsCredentialedFrontendRequests(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected allow origin for frontend, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("expected credentials to be allowed, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Expose-Headers"); got != "X-Request-ID" {
		t.Fatalf("expected request ID header to be exposed, got %q", got)
	}
}

func TestCORSMiddlewareHandlesPreflight(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not run for preflight requests")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/oauth/zetrova/authorize", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected allow origin for frontend, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("expected allowed methods header")
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Fatal("expected allowed headers header")
	}
}
