package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/afrisinc/assets/pkg/response"
)

// Recover catches any panic in downstream handlers, logs the stack trace,
// and returns a 500 instead of crashing the process.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered",
					"panic", rec,
					"stack", string(debug.Stack()),
					"path", r.URL.Path,
				)
				response.Error(w, http.StatusInternalServerError, "an unexpected error occurred")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
