package server

import (
	"errors"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/ubcesports/memberships/internal/auth"
)

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, metadata := auth.WithRequestMetadata(r.Context())
		r = r.WithContext(ctx)

		requestID := middleware.GetReqID(r.Context())
		w.Header().Set("X-Request-ID", requestID) // put request id in header and logs so it can be traced in logs

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		started := time.Now()

		next.ServeHTTP(ww, r)

		// avoid calls to /health as that will just crowd logs with no useful info
		if r.URL.Path == "/health" {
			return
		}

		status := ww.Status()
		if status == 0 {
			status = http.StatusOK
		}

		attrs := []any{
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"duration_ms", float64(time.Since(started).Microseconds()) / 1000,
			"response_size_bytes", ww.BytesWritten(),
		}
		attrs = append(attrs, clientIPAttrs(r)...)

		if metadata.UserID != "" {
			attrs = append(attrs, "user_id", metadata.UserID)
		}

		switch {
		case status >= 500:
			slog.ErrorContext(r.Context(), "request completed", attrs...)
		case status >= 400:
			slog.WarnContext(r.Context(), "request completed", attrs...)
		default:
			slog.InfoContext(r.Context(), "request completed", attrs...)
		}
	})
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			value := recover()
			if value == nil {
				return
			}

			if err, ok := value.(error); ok && errors.Is(err, http.ErrAbortHandler) {
				panic(value)
			}

			attrs := []any{
				"request_id", middleware.GetReqID(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"panic", value,
				"stack", string(debug.Stack()),
			}
			attrs = append(attrs, clientIPAttrs(r)...)

			if metadata := auth.RequestMetadataFromContext(r.Context()); metadata != nil && metadata.UserID != "" {
				attrs = append(attrs, "user_id", metadata.UserID)
			}

			slog.ErrorContext(r.Context(), "panic recovered", attrs...)

			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}()

		next.ServeHTTP(w, r)
	})
}

// remote_ip is the TCP peer address and can't be spoofed. forwarded_for is the raw
// X-Forwarded-For header, which is only trustworthy when set by our own reverse proxy
func clientIPAttrs(r *http.Request) []any {
	remoteIP := r.RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		remoteIP = host
	}

	attrs := []any{"remote_ip", remoteIP}
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		attrs = append(attrs, "forwarded_for", forwardedFor)
	}
	return attrs
}
