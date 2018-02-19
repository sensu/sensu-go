package routers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// HandlersRouter handles requests for /handlers
type HandlersRouter struct {
	controller actions.HandlerController
}

// NewHandlersRouter instantiates new router for controlling handler resources
func NewHandlersRouter(store store.HandlerStore) *HandlersRouter {
	return &HandlersRouter{
		controller: actions.NewHandlerController(store),
	}
}

// Mount the HandlersRouter to a parent Router
func (r *HandlersRouter) Mount(parent *mux.Router) {
	routes := resourceRoute{router: parent, pathPrefix: "/handlers"}
	routes.create(r.create)
	routes.destroy(r.destroy)
	routes.index(r.list)
	routes.show(r.find)
	routes.update(r.update)

}

func (r *HandlersRouter) create(req *http.Request) (interface{}, error) {
	handler := types.Handler{}
	if err := unmarshalBody(req, &handler); err != nil {
		return nil, err
	}

	return handler, r.controller.Create(req.Context(), handler)
}

func (r *HandlersRouter) destroy(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	return nil, r.controller.Destroy(req.Context(), id)
}

func (r *HandlersRouter) find(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	return r.controller.Find(req.Context(), id)
}

func (r *HandlersRouter) list(req *http.Request) (interface{}, error) {
	return r.controller.Query(req.Context())
}

func (r *HandlersRouter) update(req *http.Request) (interface{}, error) {
	handler := types.Handler{}
	if err := unmarshalBody(req, &handler); err != nil {
		return nil, err
	}

	return handler, r.controller.Update(req.Context(), handler)
}
