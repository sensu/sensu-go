package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// OrganizationsRouter handles requests for /organizations
type OrganizationsRouter struct {
	controller actions.OrganizationsController
}

// NewOrganizationsRouter instantiates new router for controlling check resources
func NewOrganizationsRouter(store store.OrganizationStore) *OrganizationsRouter {
	return &OrganizationsRouter{
		controller: actions.NewOrganizationsController(store),
	}
}

// Mount the OrganizationsRouter to a parent Router
func (r *OrganizationsRouter) Mount(parent *mux.Router) {
	routes := resourceRoute{router: parent, pathPrefix: "/organizations"}
	routes.index(r.list)
	routes.show(r.find)
	routes.create(r.create)
	routes.update(r.update)
	routes.destroy(r.destroy)
}

func (r *OrganizationsRouter) list(req *http.Request) (interface{}, error) {
	params := actions.QueryParams(mux.Vars(req))
	records, err := r.controller.Query(req.Context(), params)
	return records, err
}

func (r *OrganizationsRouter) find(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	record, err := r.controller.Find(req.Context(), params["id"])
	return record, err
}

func (r *OrganizationsRouter) create(req *http.Request) (interface{}, error) {
	org := types.Organization{}
	if err := unmarshalBody(req, &org); err != nil {
		return nil, err
	}

	err := r.controller.Create(req.Context(), org)
	return org, err
}

func (r *OrganizationsRouter) update(req *http.Request) (interface{}, error) {
	org := types.Organization{}
	if err := unmarshalBody(req, &org); err != nil {
		return nil, err
	}

	err := r.controller.Update(req.Context(), org)
	return org, err
}

func (r *OrganizationsRouter) destroy(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	err := r.controller.Destroy(req.Context(), params["id"])
	return nil, err
}
