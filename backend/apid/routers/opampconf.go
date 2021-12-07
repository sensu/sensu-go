package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
)

type OpampAgentConfRouter struct {
	store store.OpampStore
}

func NewOpampAgentConfRouter(s store.OpampStore) *OpampAgentConfRouter {
	return &OpampAgentConfRouter{
		store: s,
	}
}

func (r *OpampAgentConfRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: corev3.OpampAgentConfigResource,
	}
	routes.Path("", r.get).Methods(http.MethodGet)
	routes.Path("", r.update).Methods(http.MethodPut)
}

func (r *OpampAgentConfRouter) get(req *http.Request) (interface{}, error) {
	return r.store.GetAgentConfig(req.Context())
}

func (r *OpampAgentConfRouter) update(req *http.Request) (interface{}, error) {
	obj := &corev3.OpampAgentConfig{}
	obj.URIPath()
	if err := UnmarshalBody(req, &obj); err != nil {
		return nil, err
	}
	err := r.store.UpdateAgentConfig(req.Context(), obj)
	return obj, err
}
