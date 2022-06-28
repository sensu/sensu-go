package handlers

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// CreateV3Resource creates the resource given in the request body but only if it
// does not already exist
func (h Handlers) CreateV3Resource(r *http.Request) (interface{}, error) {
	payload := reflect.New(reflect.TypeOf(h.V3Resource).Elem())
	if err := json.NewDecoder(r.Body).Decode(payload.Interface()); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	if err := CheckV3Meta(payload.Interface(), mux.Vars(r), "id"); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	resource, ok := payload.Interface().(corev3.Resource)
	if !ok {
		return nil, actions.NewErrorf(actions.InvalidArgument)
	}

	meta := resource.GetMetadata()
	if claims := jwt.GetClaimsFromContext(r.Context()); claims != nil {
		meta.CreatedBy = claims.StandardClaims.Subject
	}

	req := storev2.NewResourceRequestFromResource(resource)
	wrapper, err := storev2.WrapResource(resource)
	if err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	if err := h.StoreV2.CreateIfNotExists(r.Context(), req, wrapper); err != nil {
		switch err := err.(type) {
		case *store.ErrAlreadyExists:
			return nil, actions.NewErrorf(actions.AlreadyExistsErr)
		case *store.ErrNotValid:
			return nil, actions.NewError(actions.InvalidArgument, err)
		default:
			return nil, actions.NewError(actions.InternalErr, err)
		}
	}

	return nil, nil
}
