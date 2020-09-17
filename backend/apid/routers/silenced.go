package routers

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/store"
)

// SilencedRouter handles requests for /users
type SilencedRouter struct {
	controller silencedController
	handlers   handlers.Handlers
}

// silencedController represents the controller needs of the SilencedRouter.
type silencedController interface {
	List(ctx context.Context, sub, check string) ([]*corev2.Silenced, error)
}

// NewSilencedRouter instantiates new router for controlling user resources
func NewSilencedRouter(store store.Store) *SilencedRouter {
	return &SilencedRouter{
		controller: actions.NewSilencedController(store),
		handlers: handlers.Handlers{
			Resource: &corev2.Silenced{},
			Store:    store,
		},
	}
}

// Mount the SilencedRouter to a parent Router
func (r *SilencedRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/namespaces/{namespace}/{resource:silenced}",
	}

	routes.Del(r.handlers.DeleteResource)
	routes.Get(r.handlers.GetResource)
	routes.Patch(r.handlers.PatchResource)
	routes.Post(r.handlers.CreateResource)
	routes.Put(r.handlers.CreateOrUpdateResource)
	routes.List(r.handlers.ListResources, corev2.SilencedFields)
	routes.ListAllNamespaces(r.handlers.ListResources, "/{resource:silenced}", corev2.SilencedFields)

	// Custom routes for listing by subscription and checks for a specific
	// namespace, in addition to all namespaces for checks.
	routes.Router.HandleFunc("/{resource:silenced}/checks/{check}", listHandler(r.list)).Methods(http.MethodGet)
	routes.Router.HandleFunc(routes.PathPrefix+"/subscriptions/{subscription}", listHandler(r.list)).Methods(http.MethodGet)
	routes.Router.HandleFunc(routes.PathPrefix+"/checks/{check}", listHandler(r.list)).Methods(http.MethodGet)
}

func (r *SilencedRouter) list(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	return r.controller.List(req.Context(), params["subscription"], params["check"])
}
