package middlewares

import (
	"net/http"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
)

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
