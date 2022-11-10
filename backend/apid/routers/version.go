package routers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
)

// VersionController represents the controller needs of the VersionRouter
type VersionController interface {
	GetVersion(ctx context.Context) *corev2.Version
}

// VersionRouter handles requests for /version
type VersionRouter struct {
	controller VersionController
}

// NewVersionRouter instantiates new router for controlling version information
func NewVersionRouter(ctrl VersionController) *VersionRouter {
	return &VersionRouter{
		controller: ctrl,
	}
}

// Mount the VersionRouter to a parent Router
func (r *VersionRouter) Mount(parent *mux.Router) {
	parent.HandleFunc("/version", r.version).Methods(http.MethodGet)
}

func (r *VersionRouter) version(w http.ResponseWriter, _ *http.Request) {
	version := r.controller.GetVersion(context.Background())
	_ = json.NewEncoder(w).Encode(version)
}
