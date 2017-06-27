package middlewares

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/sensu/sensu-go/backend/store"
)

const (
	// OrganizationKey contains the key name to retrieve the org from context
	OrganizationKey = "org"
)

// ValidateOrganization retrieves any organization passed as a query parameter
// and validates its existence against the data store
func ValidateOrganization(next http.Handler, store store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		org := r.URL.Query().Get("org")
		if org == "" {
			// If we have no organization provided, continue with the request
			next.ServeHTTP(w, r)
			return
		}

		// Verify that the organization exist
		if _, err := store.GetOrganizationByName(org); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Set the organization into the request context
		context.Set(r, OrganizationKey, org)

		next.ServeHTTP(w, r)
		return
	})
}
