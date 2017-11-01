package middlewares

import (
	"net/http"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// Authentication is a HTTP middleware that enforces authentication
type Authentication struct{}

// Then middleware
func (a Authentication) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := jwt.ExtractBearerToken(r)
		if tokenString != "" {
			token, err := jwt.ValidateToken(tokenString)
			if err != nil {
				logger.Warn("invalid token: " + err.Error())
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
		return
	})
}

// BasicAuthentication is HTTP middleware for basic authentication
func BasicAuthentication(next http.Handler, store store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		username, password, ok := r.BasicAuth()
		if !ok {
			http.Error(w, "Request unauthorized", http.StatusUnauthorized)
			return
		}

		// Authenticate against the provider
		_, err := store.AuthenticateUser(username, password)
		if err != nil {
			logger.WithField(
				"user", username,
			).Errorf("invalid username and/or password: %s", err.Error())
			http.Error(w, "Request unauthorized", http.StatusUnauthorized)
			return
		}
		// TODO: eventually break out authroization details in context from jwt claims; in this method they are too tightly bound
		claims, _ := jwt.NewClaims(username)
		ctx := jwt.SetClaimsIntoContext(r, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
		return
	})
}
