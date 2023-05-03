package middlewares

import (
	"net/http"
	"strings"

	"github.com/sensu/sensu-go/backend/apid/request"
	"github.com/sensu/sensu-go/backend/selector"
)

type Selectors struct{}

func (s Selectors) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()

		query := r.URL.Query()
		// Determine if we have a label selector
		var labelSelector *selector.Selector
		requirements := strings.Join(query["labelSelector"], " && ")
		if requirements != "" {
			var err error
			labelSelector, err = selector.ParseLabelSelector(requirements)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"message": "failed to parse labelSelector"}`))
				return
			}
		}

		// Determine if we have a field selector
		var fieldSelector *selector.Selector
		requirements = strings.Join(query["fieldSelector"], " && ")
		if requirements != "" {
			var err error
			fieldSelector, err = selector.ParseFieldSelector(requirements)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"message": "failed to parse fieldSelector"}`))
				return
			}
		}

		ctx = request.ContextWithSelector(ctx, selector.Merge(labelSelector, fieldSelector))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
