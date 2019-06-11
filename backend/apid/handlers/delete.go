package handlers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
)

// DeleteResource deletes the resources identified in the request path
func (h Handlers) DeleteResource(r *http.Request) (interface{}, error) {
	params := mux.Vars(r)
	name, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	if err := h.Store.DeleteResource(r.Context(), h.Resource.StorePrefix(), name); err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return nil, actions.NewErrorf(actions.NotFound)
		default:
			return nil, actions.NewError(actions.InternalErr, err)
		}
	}

	return nil, nil
}
