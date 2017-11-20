package middlewares

import (
	"net/http"
)

// LimitRequest is an HTTP middleware that enforces request limits
type LimitRequest struct{}

// Then middleware
func (l LimitRequest) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 512000)
		next.ServeHTTP(w, r)
	})
}
