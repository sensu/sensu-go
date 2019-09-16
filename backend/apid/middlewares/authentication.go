package middlewares

import (
	"context"
	"net/http"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
)

// BasicAuthentication is a public function that returns the HTTP middleware for
// handling basic authentication in agentd
var BasicAuthentication = basicAuthentication

// AuthStore specifies the storage requirements for auth types.
type AuthStore interface {
	// AuthenticateUser attempts to authenticate a user with the given username
	// and hashed password. An error is returned if the user does not exist, is
	// disabled or the given password does not match.
	AuthenticateUser(ctx context.Context, user, pass string) (*corev2.User, error)

	// GetUser directly retrieves a user with the given username. An error is
	// returned if the user does not exist or is disabled
	GetUser(ctx context.Context, username string) (*corev2.User, error)
}

// Authentication is a HTTP middleware that enforces authentication
type Authentication struct {
	// IgnoreUnauthorized configures the middleware to continue the handler chain
	// in the case where an access token was not present.
	IgnoreUnauthorized bool
}

// Then middleware
func (a Authentication) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tokenString := jwt.ExtractBearerToken(r)
		if tokenString != "" {
			token, err := jwt.ValidateToken(tokenString)
			if err != nil {
				logger.WithError(err).Warn("invalid token")
				writeErr(w, actions.NewErrorf(actions.Unauthenticated, "invalid credentials"))
				return
			}

			// Set the claims into the request context
			ctx = jwt.SetClaimsIntoContext(r, token.Claims.(*corev2.Claims))
			next.ServeHTTP(w, r.WithContext(ctx))

			return
		}

		// The user is not authenticated
		if a.IgnoreUnauthorized {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		writeErr(w, actions.NewErrorf(actions.Unauthenticated, "bad credentials"))
	})
}

// basicAuthentication is HTTP middleware for basic authentication
func basicAuthentication(next http.Handler, store AuthStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			writeErr(w, actions.NewErrorf(actions.Unauthenticated, "missing credentials"))
			return
		}

		// Authenticate against the provider
		user, err := store.AuthenticateUser(r.Context(), username, password)
		if err != nil {
			logger.
				WithField("user", username).
				WithError(err).
				Error("invalid username and/or password")
			writeErr(w, actions.NewErrorf(actions.Unauthenticated, "bad credentials"))
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
