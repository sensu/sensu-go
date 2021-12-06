package routers

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
)

type OpampAgentConfController interface {
	CreateOrUpdate(context.Context, *corev3.OpampAgentConfig) error
	Get(context.Context) (*corev3.OpampAgentConfig, error)
}

type OpampAgentConfRouter struct {
	controller OpampAgentConfController
}

func NewOpampAgentConfRouter(ctrl OpampAgentConfController) *OpampAgentConfRouter {
	return &OpampAgentConfRouter{
		controller: ctrl,
	}
}

func (r *OpampAgentConfRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: corev3.OpampAgentConfigResource,
	}
	routes.Path("", r.get).Methods(http.MethodGet)
	routes.Path("", r.createOrUpdate).Methods(http.MethodPut)
}

func (r *OpampAgentConfRouter) get(req *http.Request) (interface{}, error) {
	return r.controller.Get(req.Context())
}

func (r *OpampAgentConfRouter) createOrUpdate(req *http.Request) (interface{}, error) {
	obj := &corev3.OpampAgentConfig{}
	obj.URIPath()
	if err := UnmarshalBody(req, &obj); err != nil {
		return nil, err
	}
	err := r.controller.CreateOrUpdate(req.Context(), obj)
	return obj, err
}
