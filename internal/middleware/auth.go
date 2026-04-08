package middleware

import (
	"net/http"
	"strings"

	"github.com/afrisinc/assets/pkg/response"
)

// APIKeyAuth validates a static API key supplied as a Bearer token or via
// the X-API-Key header.  Swap this for JWT/OAuth middleware when ready.
func APIKeyAuth(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			provided := extractKey(r)
			if provided == "" || provided != key {
				response.Unauthorized(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func extractKey(r *http.Request) string {
	// Try Authorization: Bearer <key>
	if auth := r.Header.Get("Authorization"); auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			return parts[1]
		}
	}
	// Try X-API-Key header
	if key := r.Header.Get("X-API-Key"); key != "" {
		return key
	}
	return ""
}
