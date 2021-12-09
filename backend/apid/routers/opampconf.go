package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
)

type OpampAgentConfRouter struct {
	store store.OpampStore
	bus   messaging.MessageBus
}

func NewOpampAgentConfRouter(s store.OpampStore, bus messaging.MessageBus) *OpampAgentConfRouter {
	return &OpampAgentConfRouter{
		store: s,
		bus:   bus,
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
	if err != nil {
		return obj, err
	}
	r.bus.Publish(messaging.TopicOpampAgentConfig, obj)
	return obj, err
}
