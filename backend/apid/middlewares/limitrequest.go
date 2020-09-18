package middlewares

import (
	"io"
	"net/http"
)

// MaxBytesLimit is the default max http request size, in bytes (see https://docs.sensu.io/sensu-go/latest/api/#request-size-limit)
const (
	MaxBytesLimit = 512000
)

// LimitRequest is an HTTP middleware that enforces request limits
type LimitRequest struct {
	Limit int64
}

// Then middleware
func (l LimitRequest) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, l.Limit)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		err := r.ParseForm()
		if err != nil && err != io.EOF {
			http.Error(w, "Request exceeded max length", http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r)
	})
}
