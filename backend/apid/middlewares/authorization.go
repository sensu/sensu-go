package middlewares

import (
	"context"
	"net/http"

	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/transport"
	corev2 "github.com/sensu/sensu-go/types"
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

// BasicAuthorization performs basic authorization for entity creation via the agent websocket.
func BasicAuthorization(next http.Handler, store store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		namespace := r.Header.Get(transport.HeaderKeyNamespace)
		ctx := r.Context()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, namespace)

		user, err := store.GetUser(ctx, r.Header.Get(transport.HeaderKeyUser))
		if err != nil {
			writeErr(w, actions.NewErrorf(actions.PermissionDenied, "invalid user"))
			return
		}
		attrs := &authorization.Attributes{
			APIGroup:     "core",
			APIVersion:   "v2",
			Namespace:    namespace,
			Resource:     "events",
			ResourceName: r.Header.Get(transport.HeaderKeyAgentName),
			Verb:         "create",
			User:         *user,
		}

		auth := &rbac.Authorizer{
			Store: store,
		}
		authorized, err := auth.Authorize(ctx, attrs)
		if err != nil {
			writeErr(w, actions.NewErrorf(actions.PermissionDenied, "error authorizing session"))
			return
		}
		if !authorized {
			writeErr(w, actions.NewErrorf(actions.PermissionDenied, "session is unauthorized"))
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
