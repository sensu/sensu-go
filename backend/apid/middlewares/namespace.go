package middlewares

import (
	"context"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/types"
)

const (
	defaultNamespace = "default"
)

// Namespace retrieves the namespace passed as a query parameter and add it into
// the context of the request
type Namespace struct{}

// Then middleware
func (n Namespace) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)

		// Check if we have a namespace path variable
		namespace, err := url.PathUnescape(vars["namespace"])
		if err == nil && namespace != "" {
			// Inject the namespace into the context of the request
			ctx = context.WithValue(ctx, types.NamespaceKey, namespace)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
