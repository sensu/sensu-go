package routers

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

// ClusterIDController represents the controller needs of the ClusterIDRouter.
type ClusterIDController interface {
	Get(context.Context) (string, error)
}

// ClusterIDRouter handles requests for /id.
type ClusterIDRouter struct {
	controller ClusterIDController
}

// NewClusterIDRouter instantiates a new router for the sensu cluster id.
func NewClusterIDRouter(ctrl ClusterIDController) *ClusterIDRouter {
	return &ClusterIDRouter{
		controller: ctrl,
	}
}

// Mount the ClusterIDRouter on the given parent Router
func (r *ClusterIDRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/id",
	}

	routes.Path("", r.get).Methods(http.MethodGet)
}

func (r *ClusterIDRouter) get(req *http.Request) (interface{}, error) {
	return r.controller.Get(req.Context())
}
