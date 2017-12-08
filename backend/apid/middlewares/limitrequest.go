package middlewares

import (
	"io"
	"net/http"
)

// MaxBytesLimit is the max http request size, in bytes (see https://github.com/sensu/sensu-alpha-documentation/blob/master/97-FAQ.md)
const (
	MaxBytesLimit = 512000
)

// LimitRequest is an HTTP middleware that enforces request limits
type LimitRequest struct{}

// Then middleware
func (l LimitRequest) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, MaxBytesLimit)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		err := r.ParseForm()
		if err != nil && err != io.EOF {
			http.Error(w, "Request exceeded max length", http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r)
	})
}
