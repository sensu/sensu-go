package routers

import (
	"context"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/types"
)

type OrganizationsController interface {
	Query(ctx context.Context) ([]*types.Organization, error)
	Find(ctx context.Context, name string) (*types.Organization, error)
	Create(ctx context.Context, newOrg types.Organization) error
	CreateOrReplace(ctx context.Context, newOrg types.Organization) error
	Update(ctx context.Context, given types.Organization) error
	Destroy(ctx context.Context, name string) error
}

// OrganizationsRouter handles requests for /organizations
type OrganizationsRouter struct {
	controller OrganizationsController
}

// NewOrganizationsRouter instantiates new router for controlling check resources
func NewOrganizationsRouter(ctrl OrganizationsController) *OrganizationsRouter {
	return &OrganizationsRouter{
		controller: ctrl,
	}
}

// Mount the OrganizationsRouter to a parent Router
func (r *OrganizationsRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{Router: parent, PathPrefix: "/rbac/organizations"}
	routes.GetAll(r.list)
	routes.Get(r.find)
	routes.Post(r.create)
	routes.Del(r.destroy)
	routes.Put(r.createOrReplace)
}

func (r *OrganizationsRouter) list(req *http.Request) (interface{}, error) {
	records, err := r.controller.Query(req.Context())
	return records, err
}

func (r *OrganizationsRouter) find(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	record, err := r.controller.Find(req.Context(), id)
	return record, err
}

func (r *OrganizationsRouter) create(req *http.Request) (interface{}, error) {
	org := types.Organization{}
	if err := UnmarshalBody(req, &org); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), org)
	return org, err
}

func (r *OrganizationsRouter) createOrReplace(req *http.Request) (interface{}, error) {
	org := types.Organization{}
	if err := UnmarshalBody(req, &org); err != nil {
		return nil, err
	}

	err := r.controller.CreateOrReplace(req.Context(), org)
	return org, err
}

func (r *OrganizationsRouter) destroy(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	err = r.controller.Destroy(req.Context(), id)
	return nil, err
}
