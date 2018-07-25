package routers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/types"
)

type statusFn func() types.StatusMap

type HealthController interface {
	Health(ctx context.Context) []*types.ClusterHealth
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
	respMap := r.controller.Health(context.Background())
	fmt.Println(respMap)
	fmt.Println(r.status().Healthy())
	if !r.status().Healthy() {
		http.Error(w, "", http.StatusServiceUnavailable)
		return
	}
	/*
		resp, err := etcdHealth
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(resp)
	*/
	// return a body with cluster health info from etcd.Healthy()
	// Implicitly returns 200
}
