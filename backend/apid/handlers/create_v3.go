package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// CreateV3Resource creates the resource given in the request body but only if it
// does not already exist
func (h Handlers[R, T]) CreateResource(r *http.Request) (corev3.Resource, error) {
	var payload R
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	meta := payload.GetMetadata()
	if meta == nil {
		return nil, actions.NewError(actions.InvalidArgument, errors.New("nil metadata"))
	}
	if err := checkMeta(*meta, mux.Vars(r), "id"); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	if claims := jwt.GetClaimsFromContext(r.Context()); claims != nil {
		meta.CreatedBy = claims.StandardClaims.Subject
	}

	gstore := storev2.Of[R](h.Store)

	if err := gstore.CreateIfNotExists(r.Context(), payload); err != nil {
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
