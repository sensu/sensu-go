package routers

import (
	"context"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/api"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
)

// NamespacesRouter handles requests for /namespaces
type NamespacesRouter struct {
	handlers handlers.Handlers
	store    store.ResourceStore
	auth     authorization.Authorizer
}

// NewNamespacesRouter instantiates new router for controlling check resources
func NewNamespacesRouter(store store.ResourceStore, auth authorization.Authorizer) *NamespacesRouter {
	return &NamespacesRouter{
		store: store,
		auth:  auth,
		handlers: handlers.Handlers{
			Resource: &corev2.Namespace{},
			Store:    store,
		},
	}
}

// Mount the NamespacesRouter to a parent Router
func (r *NamespacesRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/{resource:namespaces}",
	}

	routes.Del(r.handlers.DeleteResource)
	routes.Get(r.handlers.GetResource)
	routes.List(r.list, corev2.NamespaceFields)
	routes.Post(r.handlers.CreateResource)
	routes.Put(r.handlers.CreateOrUpdateResource)
}

func (r *NamespacesRouter) list(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	client := api.NewNamespaceClient(r.store, r.auth)
	namespaces, err := client.ListNamespaces(ctx, pred)
	if err != nil {
		return nil, err
	}
	result := make([]corev2.Resource, len(namespaces))
	for i := range namespaces {
		result[i] = namespaces[i]
	}
	return result, nil
}
