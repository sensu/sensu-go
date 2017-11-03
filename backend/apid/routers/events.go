package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/useractions"
	"github.com/sensu/sensu-go/backend/store"
)

// EventsRouter handles requests for /events
type EventsRouter struct {
	controller useractions.EventActions
}

// NewEventsRouter instantiates new events controller
func NewEventsRouter(store store.EventStore) *EventsRouter {
	return &EventsRouter{
		controller: useractions.NewEventActions(nil, store),
	}
}

// Mount the EventsRouter to a parent Router
func (r *EventsRouter) Mount(parent *mux.Router) {
	routes := resourceRoute{router: parent, pathPrefix: "/events"}
	routes.index(r.list)
	routes.path("{entity}", r.listByEntity).Methods(http.MethodGet)
	routes.path("{entity}/{check}", r.find).Methods(http.MethodGet)
}

func (r *EventsRouter) find(req *http.Request) (interface{}, error) {
	fetcher := r.controller.WithContext(req.Context())
	params := useractions.QueryParams(mux.Vars(req))
	record, err := fetcher.Find(params)
	return record, err
}

func (r *EventsRouter) list(req *http.Request) (interface{}, error) {
	fetcher := r.controller.WithContext(req.Context())
	records, err := fetcher.Query(useractions.QueryParams{})
	return records, err
}

func (r *EventsRouter) listByEntity(req *http.Request) (interface{}, error) {
	fetcher := r.controller.WithContext(req.Context())
	params := useractions.QueryParams(mux.Vars(req))
	records, err := fetcher.Query(params)
	return records, err
}
