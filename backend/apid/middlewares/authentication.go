package middlewares

import (
	"context"
	"net/http"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/types"
)

// AuthStore specifies the storage requirements for auth types.
type AuthStore interface {
	// AuthenticateUser attempts to authenticate a user with the given username
	// and hashed password. An error is returned if the user does not exist, is
	// disabled or the given password does not match.
	AuthenticateUser(ctx context.Context, user, pass string) (*types.User, error)
}

// Authentication is a HTTP middleware that enforces authentication
type Authentication struct{}

// Then middleware
func (a Authentication) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := jwt.ExtractBearerToken(r)
		if tokenString != "" {
			token, err := jwt.ValidateToken(tokenString)
			if err != nil {
				logger.WithError(err).Warn("invalid token")
				http.Error(w, "Invalid token given", http.StatusUnauthorized)
				return
			}

			// Set the claims into the request context
			ctx := jwt.SetClaimsIntoContext(r, token.Claims.(*types.Claims))

			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// The user is not authenticated
		http.Error(w, "Bad credentials given", http.StatusUnauthorized)
	})
}

// BasicAuthentication is HTTP middleware for basic authentication
func BasicAuthentication(next http.Handler, store AuthStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			http.Error(w, "Request unauthorized", http.StatusUnauthorized)
			return
		}

		// Authenticate against the provider
		_, err := store.AuthenticateUser(r.Context(), username, password)
		if err != nil {
			logger.WithField(
				"user", username,
			).WithError(err).Errorf("invalid username and/or password")
			http.Error(w, "Request unauthorized", http.StatusUnauthorized)
			return
		}
		// TODO: eventually break out authroization details in context from jwt claims; in this method they are too tightly bound
		claims, _ := jwt.NewClaims(username)
		ctx := jwt.SetClaimsIntoContext(r, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
