package middlewares

import (
	"context"
	"net/http"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// Authorization is an HTTP middleware that enforces authorization
func Authorization(next http.Handler, store store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := jwt.GetClaimsFromContext(r)
		if claims == nil {
			http.Error(w, "No claims found for JWT", http.StatusInternalServerError)
			return
		}

		roles, err := store.GetRoles()
		if err != nil {
			http.Error(w, "Error fetching roles from store", http.StatusInternalServerError)
		}

		user, err := store.GetUser(claims.StandardClaims.Subject)
		if err != nil {
			http.Error(w, "Error fetching user from store", http.StatusInternalServerError)
		}

		userRoles := []*types.Role{}

		for _, userRoleName := range user.Roles {
			// TODO: (JK) we're not protecting against cases where a
			// userRoleName doesn't actually have a corresponding role
			for _, role := range roles {
				if userRoleName == role.Name {
					userRoles = append(userRoles, role)
					break
				}
			}
		}

		ctx := context.WithValue(r.Context(), "roles", userRoles)
		next.ServeHTTP(w, r.WithContext(ctx))
		return
	})
}
