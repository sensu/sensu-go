package middlewares

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
)

// AllowList verifies that the access token provided is authorized
type AllowList struct {
	Store store.Store

	// IgnoreMissingClaims configures the middleware to continue the handler chain
	// in the case where an access token was not present.
	IgnoreMissingClaims bool
}

// Then ...
func (m AllowList) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := jwt.GetClaimsFromContext(r.Context())
		if claims == nil {
			if m.IgnoreMissingClaims {
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, "Request unauthorized", http.StatusUnauthorized)
			}
			return
		}

		// Validate that the JWT is authorized
		if _, err := m.Store.GetToken(claims.Subject, claims.Id); err != nil {
			logger.WithFields(logrus.Fields{
				"user":         claims.Subject,
				"access token": claims.Id,
			}).WithError(err).Error("access token is not authorized")
			http.Error(w, "Request unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
