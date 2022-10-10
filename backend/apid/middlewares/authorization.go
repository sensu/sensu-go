package middlewares

import (
	"net/http"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
)

// Authorization is an HTTP middleware that enforces authorization
type Authorization struct {
	Authorizer authorization.Authorizer
}

func namespaceGetAttrs(attrs *authorization.Attributes) bool {
	return (attrs.APIGroup == "core" &&
		attrs.APIVersion == "v2" &&
		attrs.Resource == (&corev2.Namespace{}).RBACName() &&
		(attrs.Verb == "get" || attrs.Verb == "list"))
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

		if namespaceGetAttrs(attrs) {
			// Special case for getting namespaces - it is up to the router to handle authz
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		authorized, err := a.Authorizer.Authorize(ctx, attrs)
		if err != nil {
			if _, ok := err.(rbac.ErrRoleNotFound); ok {
				writeErr(w, actions.NewErrorf(
					actions.PermissionDenied,
					err.Error(),
				))
				return
			}
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
