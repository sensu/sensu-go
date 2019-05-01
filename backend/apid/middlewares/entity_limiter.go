package middlewares

import (
	"net/http"
)

// EntityLimiter ...
type EntityLimiter struct{}

// Then middleware
func (e EntityLimiter) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
