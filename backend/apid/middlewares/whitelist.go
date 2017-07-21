package middlewares

import (
	"net/http"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
)

// Whitelist verifies if the access token provided is whitelisted
func Whitelist(next http.Handler, store store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := jwt.GetClaimsFromContext(r)
		if claims == nil {
			http.Error(w, "Request unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate that the JWT is whitelisted
		if _, err := store.GetToken(claims.Subject, claims.Id); err != nil {
			logger.WithField(
				"user", claims.Subject,
			).Errorf("access token %s is not authorized: %s", claims.Id, err.Error())
			http.Error(w, "Request unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
		return
	})
}
