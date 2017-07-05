package middlewares

import (
	"context"
	"net/http"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

const (
	defaultOrganization = "default"
)

// Organization retrieves any organization passed as a query parameter
// and validates its existence against the data store and then adds it to the
// request context
func Organization(next http.Handler, store store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		org := r.URL.Query().Get("org")
		if org == "" {
			org = defaultOrganization
		} else {
			// Verify that the organization exist
			if _, err := store.GetOrganizationByName(org); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		ctx := context.WithValue(r.Context(), types.OrganizationKey, org)
		next.ServeHTTP(w, r.WithContext(ctx))
		return
	})
}
