package util

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ubcesports/memberships/internal/auth"
)

func WriteJson(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func SetPublicCacheHeader(w http.ResponseWriter, maxAge int, sharedMaxAge int, staleWhileRevalidate int) {
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d, s-maxage=%d, stale-while-revalidate=%d", maxAge, sharedMaxAge, staleWhileRevalidate))
}

func WriteApiResponse(w http.ResponseWriter, status int, code, message, requestID string) {
	WriteJson(w, status, map[string]string{"code": code, "message": message, "request_id": requestID})
}

func CurrentUserID(r *http.Request) (string, bool) {
	session := auth.SessionFromContext(r.Context())
	if session == nil || session.User == nil {
		return "", false
	}
	value, ok := session.User.ID.(string)
	if !ok || value == "" {
		return "", false
	}
	return value, true
}
