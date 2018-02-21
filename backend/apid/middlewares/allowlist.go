package middlewares

import (
	"net/http"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
)

// AllowList verifies that the access token provided is authorized
type AllowList struct {
	Store store.Store
}

// Then ...
func (m AllowList) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := jwt.GetClaimsFromContext(r.Context())
		if claims == nil {
			http.Error(w, "Request unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate that the JWT is authorized
		if _, err := m.Store.GetToken(claims.Subject, claims.Id); err != nil {
			logger.WithField(
				"user", claims.Subject,
			).Errorf("access token %s is not authorized: %s", claims.Id, err.Error())
			http.Error(w, "Request unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
