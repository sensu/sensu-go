package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/useractions"
	"github.com/sensu/sensu-go/backend/store"
)

// ChecksRouter handles requests for /checks
type ChecksRouter struct {
	controller useractions.CheckActions
}

// NewChecksRouter instantiates new checks controller
func NewChecksRouter(store store.CheckConfigStore) *ChecksRouter {
	return &ChecksRouter{
		controller: useractions.NewCheckActions(nil, store),
	}
}

// Mount the ChecksRouter to a parent Router
func (r *ChecksRouter) Mount(parent *mux.Router) {
	routes := resourceRoute{router: parent, pathPrefix: "/checks"}
	routes.index(r.list)
	routes.show(r.find)
}

func (r *ChecksRouter) list(req *http.Request) (interface{}, error) {
	fetcher := r.controller.WithContext(req.Context())
	records, err := fetcher.Query(useractions.QueryParams{})
	return records, err
}

func (r *ChecksRouter) find(req *http.Request) (interface{}, error) {
	fetcher := r.controller.WithContext(req.Context())
	params := useractions.QueryParams(mux.Vars(req))
	record, err := fetcher.Find(params)
	return record, err
}
