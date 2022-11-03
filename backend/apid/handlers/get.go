package handlers

import (
	"net/http"
	"net/url"
	"reflect"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
)

// GetResource retrieves the resource identified in the request path
func (h Handlers) GetResource(r *http.Request) (interface{}, error) {
	params := mux.Vars(r)
	name, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}

	v := reflect.New(reflect.TypeOf(h.Resource).Elem())
	resource, ok := v.Interface().(corev2.Resource)
	if !ok {
		return nil, actions.NewErrorf(actions.InternalErr)
	}

	if err := h.Store.GetResource(r.Context(), name, resource); err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return nil, actions.NewErrorf(actions.NotFound)
		default:
			return nil, actions.NewError(actions.InternalErr, err)
		}
	}

	return resource, nil
}
