package routers

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// TessenController represents the controller needs of the TessenRouter.
type TessenController interface {
	CreateOrUpdate(context.Context, *corev2.TessenConfig) error
	Get(context.Context) (*corev2.TessenConfig, error)
}

// TessenRouter handles requests for /tessen.
type TessenRouter struct {
	controller TessenController
}

// NewTessenRouter instantiates a new router for tessen.
func NewTessenRouter(ctrl TessenController) *TessenRouter {
	return &TessenRouter{
		controller: ctrl,
	}
}

// Mount the TessenRouter on the given parent Router
func (r *TessenRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: corev2.TessenPath,
	}

	routes.Path("", r.get).Methods(http.MethodGet)
	routes.Path("", r.createOrUpdate).Methods(http.MethodPut)
}

func (r *TessenRouter) createOrUpdate(req *http.Request) (interface{}, error) {
	obj := &corev2.TessenConfig{}
	if err := UnmarshalBody(req, &obj); err != nil {
		return nil, err
	}

	err := r.controller.CreateOrUpdate(req.Context(), obj)
	return obj, err
}

func (r *TessenRouter) get(req *http.Request) (interface{}, error) {
	obj, err := r.controller.Get(req.Context())
	return obj, err
}
