package routers

import (
	"context"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/types"
)

// EnvironmentController represents the controller needs of the EnvironmentsRouter.
type EnvironmentController interface {
	Create(context.Context, types.Environment) error
	CreateOrReplace(context.Context, types.Environment) error
	Update(context.Context, types.Environment) error
	Query(context.Context, string) ([]*types.Environment, error)
	Find(context.Context, string, string) (*types.Environment, error)
	Destroy(context.Context, string, string) error
}

// EnvironmentsRouter handles requests for /rbac/organizations/{org}/environments
type EnvironmentsRouter struct {
	controller EnvironmentController
}

// NewEnvironmentsRouter instantiates new router for controlling check resources
func NewEnvironmentsRouter(ctrl EnvironmentController) *EnvironmentsRouter {
	return &EnvironmentsRouter{
		controller: ctrl,
	}
}

// Mount the EnvironmentsRouter to a parent Router
func (r *EnvironmentsRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{Router: parent, PathPrefix: "/rbac/organizations"}
	routes.Path("{organization}/environments", r.list).Methods(http.MethodGet)
	routes.Path("{organization}/environments/{environment}", r.find).Methods(http.MethodGet)
	routes.Path("{organization}/environments", r.create).Methods(http.MethodPost)
	routes.Path("{organization}/environments/{environment}", r.createOrReplace).Methods(http.MethodPut)
	routes.Path("{organization}/environments/{environment}", r.destroy).Methods(http.MethodDelete)
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
	vars := mux.Vars(req)
	if org := vars["organization"]; len(org) > 0 {
		env.Organization, err = url.PathUnescape(org)
		if err != nil {
			return nil, err
		}
	}

	err = r.controller.Create(req.Context(), env)
	return env, err
}

func (r *EnvironmentsRouter) createOrReplace(req *http.Request) (interface{}, error) {
	env := types.Environment{}
	var err error
	if err = unmarshalBody(req, &env); err != nil {
		return nil, err
	}
	vars := mux.Vars(req)
	if org := vars["organization"]; len(org) > 0 {
		env.Organization, err = url.PathUnescape(org)
		if err != nil {
			return nil, err
		}
	}

	err = r.controller.CreateOrReplace(req.Context(), env)
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
