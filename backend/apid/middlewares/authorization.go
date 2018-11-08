package middlewares

import (
	"net/http"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/types"
)

// Authorization is an HTTP middleware that enforces authorization
type Authorization struct {
	Authorizer authorization.Authorizer
}

// Then middleware
func (a Authorization) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get the claims about the user from the context
		claims := jwt.GetClaimsFromContext(ctx)
		if claims == nil {
			http.Error(w, "no claims found in the JWT", http.StatusInternalServerError)
			return
		}

		// Get the request info from context
		attrs := authorization.GetAttributes(ctx)
		if attrs == nil {
			http.Error(w, "could not retrieve the request info", http.StatusInternalServerError)
			return
		}

		// Add the user to our request info
		attrs.User = types.User{
			Username: claims.Subject,
			Groups:   claims.Groups,
		}

		authorized, err := a.Authorizer.Authorize(attrs)
		if err != nil {
			logger.WithError(err).Warning("unexpected error occurred during authorization")
			http.Error(w, "unexpected error occurred during authorization", http.StatusInternalServerError)
			return
		}
		if !authorized {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
