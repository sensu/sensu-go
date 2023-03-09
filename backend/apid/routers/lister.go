package routers

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/filters/fields"
	"github.com/sensu/sensu-go/backend/apid/filters/labels"
	"github.com/sensu/sensu-go/backend/apid/request"
	"github.com/sensu/sensu-go/backend/selector"
	"github.com/sensu/sensu-go/backend/store"
)

// ListControllerFunc represents a generic controller for listing resources
type ListControllerFunc func(ctx context.Context, pred *store.SelectionPredicate) ([]corev3.Resource, error)

// FieldsFunc represents the function to retrieve fields about a given resource
type FieldsFunc func(resource corev3.Resource) map[string]string

// WrapList handles pagination and selector filtering for listing resources.
func WrapList(list ListControllerFunc, fieldsFunc FieldsFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		pred := &store.SelectionPredicate{
			Continue: corev2.PageContinueFromContext(r.Context()),
			Limit:    int64(corev2.PageSizeFromContext(r.Context())),
		}

		routeVariables := actions.QueryParams(mux.Vars(r))
		if subcollection := url.PathEscape(routeVariables["subcollection"]); subcollection != "" {
			pred.Subcollection = subcollection
		}

		query := r.URL.Query()

		// Determine if we have a label selector
		var labelSelector *selector.Selector
		requirements := strings.Join(query["labelSelector"], " && ")
		if requirements != "" {
			labelSelector, err = selector.ParseLabelSelector(requirements)
			if err != nil {
				WriteError(w, actions.NewError(actions.InvalidArgument, err))
				return
			}
		}

		// Determine if we have a field selector
		var fieldSelector *selector.Selector
		requirements = strings.Join(query["fieldSelector"], " && ")
		if requirements != "" {
			fieldSelector, err = selector.ParseFieldSelector(requirements)
			if err != nil {
				WriteError(w, actions.NewError(actions.InvalidArgument, err))
				return
			}
		}

		// Fetch resources from the store and filter those until we hit the
		// requested amount of resources (limit) or there's no more resources (empty
		// continue token)
		resources := []corev3.Resource{}

		ctx := r.Context()
		ctx = request.ContextWithSelector(ctx, selector.Merge(labelSelector, fieldSelector))
		r = r.WithContext(ctx)
	StoreLoop:
		for {
			results, err := list(r.Context(), pred)
			if err != nil {
				WriteError(w, err)
				return
			}
			resources = append(resources, results...)

			// Apply the label and field selectors if available
			if labelSelector != nil {
				resources = labels.Filter(resources, labelSelector.Matches).([]corev3.Resource)
			}
			if fieldSelector != nil {
				resources = fields.Filter(resources, fieldSelector.Matches, fields.FieldsFunc(fieldsFunc)).([]corev3.Resource)
			}

			// Determine what to do based on the number of resources we currently have
			// and the store's selection predicate
			switch {
			case pred.Limit == 0:
				// No limit was specified so we can assume all resources were returned
				break StoreLoop
			case len(resources) == int(pred.Limit):
				// We have the exact number of requested resources.
				// E.g. limit == 2, and len(resources) == 2
				break StoreLoop
			case len(resources) > int(pred.Limit):
				// We have more resources than requested so we need to remove the excess
				// and update the continue token
				// E.g. limit == 2, but len(resources) == 3. This can happen if we had 1
				// resource before the last query, and this last query returned 2
				// resources that matched the selectors, so we now have a total of 3
				// resources instead of the 2 requested. Therefore, we now have to
				// remove the excess and move back the continue token so it corresponds
				// to the last resource we are returning.
				excessCount := len(resources) - int(pred.Limit)
				resources = resources[:len(resources)-excessCount]
				// TODO(eric | ck): Need postgres equiv of etcd.ComputContinueToken (entity store)
				// pred.Continue = coreEtcd.ComputeContinueToken(r.Context(), resources[len(resources)-1])
				break StoreLoop
			case pred.Continue == "":
				// The store indicated that there's no more keys available.
				break StoreLoop
			}
			// If we reach this point, it means we only fetched a specific number of
			// resources from the store and we filtered some of those, which means we
			// need to fetch the next batch of resources from the store until we
			// either have the requested number of resources or there's no more keys
			// in the store.
		}

		if pred.Continue != "" {
			encodedContinue := base64.RawURLEncoding.EncodeToString([]byte(pred.Continue))
			w.Header().Set(corev2.PaginationContinueHeader, encodedContinue)
		}

		RespondWith(w, r, resources)
	}
}
