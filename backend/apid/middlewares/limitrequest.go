package middlewares

import (
	"io"
	"net/http"
)

// LimitRequest is an HTTP middleware that enforces request limits
type LimitRequest struct{}

// Then middleware
func (l LimitRequest) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 512000)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		err := r.ParseForm()
		if err != nil && err != io.EOF {
			http.Error(w, "Request exceeded max length", http.StatusInternalServerError)
		}
		next.ServeHTTP(w, r)
		r.Body.Close()
	})
}
