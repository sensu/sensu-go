package routers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// EnvironmentsRouter handles requests for /rbac/organizations/{org}/environments
type EnvironmentsRouter struct {
	controller actions.EnvironmentController
}

// NewEnvironmentsRouter instantiates new router for controlling check resources
func NewEnvironmentsRouter(store store.EnvironmentStore) *EnvironmentsRouter {
	return &EnvironmentsRouter{
		controller: actions.NewEnvironmentController(store),
	}
}

// Mount the EnvironmentsRouter to a parent Router
func (r *EnvironmentsRouter) Mount(parent *mux.Router) {
	routes := resourceRoute{router: parent, pathPrefix: "/rbac/organizations"}
	routes.path("{organization}/environments", r.list).Methods(http.MethodGet)
	routes.path("{organization}/environments/{environment}", r.find).Methods(http.MethodGet)
	routes.path("{organization}/environments", r.create).Methods(http.MethodPost)
	routes.path("{organization}/environments/{environment}", r.update).Methods(http.MethodPatch)
	routes.path("{organization}/environments/{environment}", r.destroy).Methods(http.MethodDelete)
}

func (r *EnvironmentsRouter) list(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["organization"])
	if err != nil {
		return nil, err
	}
	records, err := r.controller.Query(req.Context(), id)
	return records, err
}

func (r *EnvironmentsRouter) find(req *http.Request) (interface{}, error) {
	p := mux.Vars(req)
	org, err := url.PathUnescape(p["organization"])
	if err != nil {
		return nil, err
	}
	env, err := url.PathUnescape(p["environment"])
	if err != nil {
		return nil, err
	}
	return r.controller.Find(req.Context(), org, env)
}

func (r *EnvironmentsRouter) create(req *http.Request) (interface{}, error) {
	env := types.Environment{}
	var err error
	if err = unmarshalBody(req, &env); err != nil {
		return nil, err
	}
	env.Organization, err = url.PathUnescape(mux.Vars(req)["organization"])
	if err != nil {
		return nil, err
	}

	err = r.controller.Create(req.Context(), env)
	return env, err
}

func (r *EnvironmentsRouter) update(req *http.Request) (interface{}, error) {
	env := types.Environment{}
	if err := unmarshalBody(req, &env); err != nil {
		return nil, err
	}

	err := r.controller.Update(req.Context(), env)
	return env, err
}

func (r *EnvironmentsRouter) destroy(req *http.Request) (interface{}, error) {
	p := mux.Vars(req)
	org, err := url.PathUnescape(p["organization"])
	if err != nil {
		return nil, err
	}
	env, err := url.PathUnescape(p["environment"])
	if err != nil {
		return nil, err
	}
	err = r.controller.Destroy(req.Context(), org, env)
	return nil, err
}
