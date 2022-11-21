package handlers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"

	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

func (h Handlers[R, T]) DeleteResource(r *http.Request) (corev3.Resource, error) {
	params := mux.Vars(r)
	name, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	ctx := r.Context()
	namespace := store.NewNamespaceFromContext(ctx)

	gstore := storev2.NewGenericStore[R](h.Store)

	if err := gstore.Delete(ctx, storev2.ID{Namespace: namespace, Name: name}); err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return nil, actions.NewErrorf(actions.NotFound)
		case *store.ErrNotValid:
			return nil, actions.NewError(actions.InvalidArgument, err)
		default:
			return nil, actions.NewError(actions.InternalErr, err)
		}
	}
	return nil, nil
}
