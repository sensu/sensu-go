package routers

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
)

// ListControllerFunc represents a generic controller for listing resources
type ListControllerFunc func(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error)

// listerFunc represents the function signature of a Lister
type listerFunc func(ListControllerFunc) http.HandlerFunc

// Lister represents the active Lister function
var Lister listerFunc

func init() {
	// Assign the core lister
	Lister = List
}

// List handles resources listing with pagination support
func List(list ListControllerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pred := &store.SelectionPredicate{
			Continue: corev2.PageContinueFromContext(r.Context()),
			Limit:    int64(corev2.PageSizeFromContext(r.Context())),
		}

		params := actions.QueryParams(mux.Vars(r))
		if subcollection := url.PathEscape(params["subcollection"]); subcollection != "" {
			pred.Subcollection = subcollection
		}

		results, err := list(r.Context(), pred)
		if err != nil {
			WriteError(w, err)
			return
		}

		if pred.Continue != "" {
			encodedContinue := base64.RawURLEncoding.EncodeToString([]byte(pred.Continue))
			w.Header().Set(corev2.PaginationContinueHeader, encodedContinue)
		}

		RespondWith(w, results)
	}
}

// We can't directly use a Lister in the mux.Router because it cannot be
// modified at runtime, which is required for sensu-enterprise-go, therefore we
// need this little wrapper
func listerHandler(list ListControllerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Lister(list).ServeHTTP(w, r)
	}
}
