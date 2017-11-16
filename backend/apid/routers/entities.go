package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
)

// EntitiesRouter handles requests for /entities
type EntitiesRouter struct {
	controller actions.EntityController
}

// NewEntitiesRouter instantiates new router for controlling entities resources
func NewEntitiesRouter(store store.EntityStore) *EntitiesRouter {
	return &EntitiesRouter{
		controller: actions.NewEntityController(store),
	}
}

// Mount the EntitiesRouter to a parent Router
func (r *EntitiesRouter) Mount(parent *mux.Router) {
	routes := resourceRoute{router: parent, pathPrefix: "/entities"}
	routes.destroy(r.destroy)
	routes.index(r.list)
	routes.show(r.find)
}

func (r *EntitiesRouter) destroy(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	err := r.controller.Destroy(req.Context(), params["id"])
	return nil, err
}

func (r *EntitiesRouter) find(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	record, err := r.controller.Find(req.Context(), params["id"])
	return record, err
}

func (r *EntitiesRouter) list(req *http.Request) (interface{}, error) {
	records, err := r.controller.Query(req.Context())
	return records, err
}
