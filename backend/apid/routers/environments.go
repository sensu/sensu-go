package routers

import (
	"net/http"

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
	p := mux.Vars(req)
	records, err := r.controller.Query(req.Context(), p["organization"])
	return records, err
}

func (r *EnvironmentsRouter) find(req *http.Request) (interface{}, error) {
	p := mux.Vars(req)
	return r.controller.Find(req.Context(), p["organization"], p["environment"])
}

func (r *EnvironmentsRouter) create(req *http.Request) (interface{}, error) {
	env := types.Environment{}
	if err := unmarshalBody(req, &env); err != nil {
		return nil, err
	}
	env.Organization = mux.Vars(req)["organization"]

	err := r.controller.Create(req.Context(), env)
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
	err := r.controller.Destroy(req.Context(), p["organization"], p["environment"])
	return nil, err
}
