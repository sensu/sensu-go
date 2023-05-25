package handlers

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/request"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// CreateOrUpdateResource creates or updates the resource given in the request
// body, regardless of whether it already exists or not
func (h Handlers[R, T]) CreateOrUpdateResource(r *http.Request) (HandlerResponse, error) {
	var response HandlerResponse
	payload, err := request.Resource[R](r)
	if err != nil {
		return response, actions.NewError(actions.InvalidArgument, err)
	}
	meta := payload.GetMetadata()

	if meta == nil {
		return response, actions.NewError(actions.InvalidArgument, errors.New("nil metadata"))
	}

	if err := checkMeta(*meta, mux.Vars(r), "id"); err != nil {
		return response, actions.NewError(actions.InvalidArgument, err)
	}

	ctx := matchHeaderContext(r)
	ctx = storev2.ContextWithTxInfo(ctx, &response.TxInfo)

	if claims := jwt.GetClaimsFromContext(ctx); claims != nil {
		meta.CreatedBy = claims.StandardClaims.Subject
	}

	gstore := storev2.Of[R](h.Store)

	if err := gstore.CreateOrUpdate(ctx, payload); err != nil {
		switch err := err.(type) {
		case *store.ErrPreconditionFailed:
			return response, actions.NewError(actions.PreconditionFailed, err)
		case *store.ErrNotValid:
			return response, actions.NewError(actions.InvalidArgument, err)
		default:
			return response, actions.NewError(actions.InternalErr, err)
		}
	}

	return response, nil
}
