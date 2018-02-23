package middlewares

import (
	"context"
	"net/http"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// Authorization is an HTTP middleware that enforces authorization
type Authorization struct {
	Store store.Store
}

// Then middleware
func (a Authorization) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims := jwt.GetClaimsFromContext(ctx)
		if claims == nil {
			http.Error(w, "No claims found for JWT", http.StatusInternalServerError)
			return
		}

		roles, err := a.Store.GetRoles(ctx)
		if err != nil {
			http.Error(w, "Error fetching roles from store", http.StatusInternalServerError)
			return
		}

		user, err := a.Store.GetUser(ctx, claims.StandardClaims.Subject)
		if err != nil {
			http.Error(w, "Error fetching user from store", http.StatusInternalServerError)
			return
		} else if user == nil {
			http.Error(w, "Unabled to find user() associated with access token", http.StatusInternalServerError)
			return
		}

		userRules := []types.Rule{}
		for _, userRoleName := range user.Roles {
			// TODO: (JK) we're not protecting against cases where a
			// userRoleName doesn't actually have a corresponding role
			for _, role := range roles {
				if userRoleName == role.Name {
					userRules = append(userRules, role.Rules...)
					break
				}
			}
		}

		actor := authorization.Actor{
			Name:  claims.Subject,
			Rules: userRules,
		}

		ctx = context.WithValue(ctx, types.AuthorizationActorKey, actor)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
