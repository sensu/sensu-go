package handlers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"

	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

func (h Handlers[R, T]) GetResource(r *http.Request) (HandlerResponse, error) {
	var response HandlerResponse

	params := mux.Vars(r)
	name, err := url.PathUnescape(params["id"])
	if err != nil {
		return response, err
	}

	ctx := matchHeaderContext(r)
	ctx = storev2.ContextWithTxInfo(ctx, &response.TxInfo)

	namespace := store.NewNamespaceFromContext(ctx)

	gstore := storev2.Of[R](h.Store)

	result, err := gstore.Get(ctx, storev2.ID{Namespace: namespace, Name: name})
	if err != nil {
		switch err := err.(type) {
		case *store.ErrPreconditionFailed:
			return response, actions.NewError(actions.PreconditionFailed, err)
		case *store.ErrNotFound:
			return response, actions.NewErrorf(actions.NotFound)
		case *store.ErrNotValid:
			return response, actions.NewError(actions.InvalidArgument, err)
		default:
			return response, actions.NewError(actions.InternalErr, err)
		}
	}
	response.Resource = result

	return response, nil
}
