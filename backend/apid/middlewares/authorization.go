package middlewares

import (
	"net/http"

	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/authorization"
)

// Authorization is an HTTP middleware that enforces authorization
type Authorization struct {
	Authorizer authorization.Authorizer
}

// Then middleware
func (a Authorization) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get the request info from context
		attrs := authorization.GetAttributes(ctx)
		if attrs == nil {
			writeErr(w, actions.NewErrorf(
				actions.InternalErr,
				"could not retrieve the request info",
			))
			return
		}

		authorized, err := a.Authorizer.Authorize(ctx, attrs)
		if err != nil {
			logger.WithError(err).Warning("unexpected error occurred during authorization")
			writeErr(w, actions.NewErrorf(
				actions.InternalErr,
				"unexpected error occurred during authorization",
			))
			return
		}
		if !authorized {
			writeErr(w, actions.NewErrorf(actions.PermissionDenied))
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
