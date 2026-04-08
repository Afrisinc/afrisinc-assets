package middleware

import "net/http"

// MaxBodySize limits the request body to n bytes before the handler reads it.
// This protects against slow-loris and oversized upload attacks at the HTTP
// layer, before multipart parsing begins.
func MaxBodySize(n int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, n)
			next.ServeHTTP(w, r)
		})
	}
}
