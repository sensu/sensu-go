package middlewares

import (
	"context"
	"net/http"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

const (
	defaultNamespace = "default"
)

// Namespace retrieves the namespace passed as a query parameter and validate
// its existence against the data store and then add them to the request context
type Namespace struct {
	Store store.Store
}

// Then middleware
func (n Namespace) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var namespace string
		if namespace = r.URL.Query().Get("namespace"); namespace == "" {
			namespace = defaultNamespace
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, types.NamespaceKey, namespace)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
