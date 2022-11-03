package routers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// HealthController represents the controller needs of the HealthRouter
type HealthController interface {
	GetClusterHealth(ctx context.Context) *corev2.HealthResponse
}

// HealthRouter handles requests for /health
type HealthRouter struct {
	controller HealthController
	mu         sync.Mutex
}

// NewHealthRouter instantiates new router for controlling health info
func NewHealthRouter(ctrl HealthController) *HealthRouter {
	return &HealthRouter{
		controller: ctrl,
	}
}

// Mount the HealthRouter to a parent Router
func (r *HealthRouter) Mount(parent *mux.Router) {
	parent.HandleFunc("/health", r.health).Methods(http.MethodGet)
}

func (r *HealthRouter) health(w http.ResponseWriter, req *http.Request) {
	timeout, err := parseTimeout(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := req.Context()
	if timeout > 0 {
		// We're storing the timeout as a value so it can be used by several
		// contexts in GetClusterHealth, which is a concurrent gatherer.
		ctx = context.WithValue(ctx, store.ContextKeyTimeout, time.Duration(timeout)*time.Second)
	}
	r.mu.Lock()
	clusterHealth := r.controller.GetClusterHealth(ctx)
	r.mu.Unlock()
	_ = json.NewEncoder(w).Encode(clusterHealth)
}

// Swap swaps the health controller of the health router.
func (r *HealthRouter) Swap(newCtl HealthController) {
	r.mu.Lock()
	r.controller = newCtl
	r.mu.Unlock()
}
