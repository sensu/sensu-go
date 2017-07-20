package middlewares

import (
	"net/http"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
)

type Authentication struct{}

// Authentication is a HTTP middleware that enforces authentication
func (a Authentication) Register(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := jwt.ExtractBearerToken(r)
		if tokenString != "" {
			token, err := jwt.ValidateToken(tokenString)
			if err == nil {
				// Set the claims into the request context
				ctx := jwt.SetClaimsIntoContext(r.Context(), token)

				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// The user is not authenticated
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
		return
	})
}
