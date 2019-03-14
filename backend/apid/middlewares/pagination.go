package middlewares

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/sensu/sensu-go/types"
)

// Pagination retrieves the "limit" and "continue" query parameters and add them
// to the request's context.
type Pagination struct{}

func (p Pagination) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)

		limit, err := strconv.Atoi(vars["limit"])
		if err != nil {
			limit = 0
		}
		ctx = context.WithValue(ctx, types.PageSizeKey, limit)

		continueToken, err := url.PathUnescape(vars["continue"])
		if err != nil {
			continueToken = ""
		}
		ctx = context.WithValue(ctx, types.PageContinueKey, continueToken)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
