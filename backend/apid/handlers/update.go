package handlers

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
)

// CreateOrUpdateResource ...
func (h Handlers) CreateOrUpdateResource(r *http.Request) (interface{}, error) {
	payload := reflect.New(reflect.TypeOf(h.Resource).Elem())
	if err := json.NewDecoder(r.Body).Decode(payload.Interface()); err != nil {
		logger.WithError(err).Error("unable to read request body")
		return nil, err
	}

	if err := checkMeta(payload.Interface(), mux.Vars(r)); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	resource, ok := payload.Interface().(corev2.Resource)
	if !ok {
		return nil, actions.NewErrorf(actions.InvalidArgument)
	}

	if err := h.Store.CreateOrUpdateResource(r.Context(), resource); err != nil {
		switch err := err.(type) {
		case *store.ErrNotValid:
			return nil, actions.NewErrorf(actions.InvalidArgument)
		default:
			return nil, actions.NewError(actions.InternalErr, err)
		}
	}

	return resource, nil
}
