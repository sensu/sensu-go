package middlewares

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/sensu/sensu-go/backend/apid/actions"
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
				writeErr(w, actions.NewErrorf(actions.Unauthenticated))
			}
			return
		} else if claims.APIKey {
			next.ServeHTTP(w, r)
			return
		}

		// Validate that the JWT is authorized
		if _, err := m.Store.GetToken(claims.Subject, claims.Id); err != nil {
			logger = logger.WithFields(logrus.Fields{
				"token_id": claims.Id,
				"user":     claims.Subject,
			})
			switch err := err.(type) {
			case *store.ErrNotFound:
				logger.WithError(err).Info("access token is unauthorized")
				writeErr(w, actions.NewErrorf(actions.Unauthenticated))
			default:
				logger.WithError(err).Error("unexpected error occurred during authorization")
				writeErr(w, actions.NewErrorf(
					actions.InternalErr,
					"unexpected error occurred during authorization",
				))
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}
