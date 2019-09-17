package agentd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/middlewares"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/transport"
)

// AuthenticationMiddleware represents the middleware used for authentication
var AuthenticationMiddleware mux.MiddlewareFunc

// AuthorizationMiddleware represents the middleware used for authorization
var AuthorizationMiddleware mux.MiddlewareFunc

// authenticate is the abstraction layer required to be able to change at
// runtime the actual function assigned to AuthenticationMiddleware above
func authenticate(next http.Handler) http.Handler {
	return AuthenticationMiddleware(next)
}

// authorize is the abstraction layer required to be able to change at
// runtime the actual function assigned to AuthenticationMiddleware above
func authorize(next http.Handler) http.Handler {
	return AuthorizationMiddleware(next)
}

// AuthStore specifies the storage requirements for authentication and
// authorization.
type AuthStore interface {
	// AuthenticateUser attempts to authenticate a user with the given username
	// and hashed password. An error is returned if the user does not exist, is
	// disabled or the given password does not match.
	AuthenticateUser(ctx context.Context, user, pass string) (*corev2.User, error)

	// GetUser directly retrieves a user with the given username. An error is
	// returned if the user does not exist or is disabled
	GetUser(ctx context.Context, username string) (*corev2.User, error)
}

type authenticationMiddleware struct {
	store AuthStore
}

// Middleware represents the core authentication middleware for agentd, which
// consists of basic authentication.
func (a *authenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			http.Error(w, "missing credentials", http.StatusUnauthorized)
			return
		}

		// Authenticate against the provider
		user, err := a.store.AuthenticateUser(r.Context(), username, password)
		if err != nil {
			logger.
				WithField("user", username).
				WithError(err).
				Error("invalid username and/or password")
			http.Error(w, "bad credentials", http.StatusUnauthorized)
			return
		}

		// The user was authenticated against the local store, therefore add the
		// system:user group so it can view itself and change its password
		user.Groups = append(user.Groups, "system:user")

		// TODO: eventually break out authorization details in context from jwt
		// claims; in this method they are too tightly bound
		claims, _ := jwt.NewClaims(user)
		ctx := jwt.SetClaimsIntoContext(r, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type authorizationMiddleware struct {
	store store.Store
}

// Middleware represents the core authorization middleware for agentd, which
// consists of making sure the agent's entity is authorized to create events in
// the given namespace
func (a *authorizationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("authorizationMiddleware")
		namespace := r.Header.Get(transport.HeaderKeyNamespace)
		ctx := r.Context()
		ctx = context.WithValue(ctx, corev2.NamespaceKey, namespace)

		attrs := &authorization.Attributes{
			APIGroup:     "core",
			APIVersion:   "v2",
			Namespace:    namespace,
			Resource:     "events",
			ResourceName: r.Header.Get(transport.HeaderKeyAgentName),
			Verb:         "create",
		}

		err := middlewares.GetUser(ctx, attrs)
		if err != nil {
			logger.WithError(err).Info("could not get user from request context")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		auth := &rbac.Authorizer{
			Store: a.store,
		}
		authorized, err := auth.Authorize(ctx, attrs)
		if err != nil {
			logger.WithError(err).Error("unexpected error while authorization the session")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		if !authorized {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
