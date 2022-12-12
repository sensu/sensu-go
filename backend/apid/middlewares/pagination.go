package middlewares

import (
	"context"
	"encoding/base64"
	"net/http"
	"strconv"

	corev2 "github.com/sensu/core/v2"
)

// Pagination retrieves the "limit" and "continue" query parameters and add them
// to the request's context.
type Pagination struct{}

func (p Pagination) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		limit, err := strconv.Atoi(r.FormValue("limit"))
		if err != nil {
			limit = 0
		}
		ctx = context.WithValue(ctx, corev2.PageSizeKey, limit)

		// Decode the continue token with base64url (RFC 4648), without padding
		continueToken := ""
		decodedContinueToken, err := base64.RawURLEncoding.DecodeString(r.FormValue("continue"))
		if err == nil {
			continueToken = string(decodedContinueToken)
		}
		ctx = context.WithValue(ctx, corev2.PageContinueKey, continueToken)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
