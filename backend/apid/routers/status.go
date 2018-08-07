package routers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/types"
)

type statusFn func() types.StatusMap

type HealthController interface {
	GetClusterHealth(ctx context.Context) *types.HealthResponse
}

// StatusRouter handles requests for /events
type StatusRouter struct {
	status     statusFn
	controller HealthController
}

// NewStatusRouter instantiates new events controller
func NewStatusRouter(status statusFn, ctrl HealthController) *StatusRouter {
	return &StatusRouter{
		status:     status,
		controller: ctrl,
	}
}

// Mount the StatusRouter to a parent Router
func (r *StatusRouter) Mount(parent *mux.Router) {
	parent.HandleFunc("/info", actionHandler(r.info)).Methods(http.MethodGet)
	parent.HandleFunc("/health", r.health).Methods(http.MethodGet)
}

func (r *StatusRouter) info(req *http.Request) (interface{}, error) {
	return r.status(), nil
}

func (r *StatusRouter) health(w http.ResponseWriter, _ *http.Request) {
	clusterHealth := r.controller.GetClusterHealth(context.Background())
	if !r.status().Healthy() {
		http.Error(w, "", http.StatusServiceUnavailable)
		return
	}

	_ = json.NewEncoder(w).Encode(clusterHealth)
}
