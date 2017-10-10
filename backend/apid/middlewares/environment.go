package middlewares

import (
	"context"
	"net/http"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

const (
	defaultEnvironment  = "default"
	defaultOrganization = "default"
)

// Environment retrieves any organization and environment passed as query
// parameters and validate their existence against the data store and then add
// them to the request context
type Environment struct {
	Store store.Store
}

// Then middleware
func (m Environment) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var env, org string
		if env = r.URL.Query().Get("env"); env == "" || env == "*" {
			env = defaultEnvironment
		}
		if org = r.URL.Query().Get("org"); org == "" || org == "*" {
			org = defaultOrganization
		}

		// Verify that the environment exist
		if _, err := m.Store.GetEnvironment(r.Context(), org, env); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, types.OrganizationKey, org)
		ctx = context.WithValue(ctx, types.EnvironmentKey, env)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
