package middlewares

import (
	"io"
	"net/http"

	"github.com/sensu/sensu-go/backend/authorization"
)

// MaxBytesLimit is the default max http request size, in bytes (see https://docs.sensu.io/sensu-go/latest/api/#request-size-limit)
const (
	UnauthedMaxBytesLimit = 8
)

// LimitRequest is an HTTP middleware that enforces request limits
type LimitUnauthedRequest struct {
	Limit int64
}

// Then middleware
func (l LimitUnauthedRequest) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		attrs := &authorization.Attributes{}
		user := GetUser(ctx, attrs)

		if user == authorization.ErrNoClaims {
			r.Body = http.MaxBytesReader(w, r.Body, l.Limit)
			r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			err := r.ParseForm()
			if err != nil && err != io.EOF {
				http.Error(w, "Request exceeded max length", http.StatusInternalServerError)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
