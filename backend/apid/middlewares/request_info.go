package middlewares

import (
	"context"
	"net/http"
)

// RequestInfo is an HTTP middleware that populates a context for the request,
// to be used by other middlewares. You probably want this middleware to be
// executed early in the middleware stack.
type RequestInfo struct{}

func (i RequestInfo) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// APIGroup, Namespace, ... anything else we can parse out of the
		// request and want to include.
		ctx = context.WithValue(ctx, "APIGroup", "value")
		ctx = context.WithValue(ctx, "Namespace", "value")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
