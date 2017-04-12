package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// HealthController defines the fields required by HealthController.
type HealthController struct {
	Store  store.Store
	Status func() types.StatusMap
}

// Register should define an association between HTTP routes and their
// respective handlers defined within this Controller.
func (c *HealthController) Register(r *mux.Router) {
	r.HandleFunc("/health", c.many).Methods(http.MethodGet)
}

// many handles requests to /info
func (c *HealthController) many(w http.ResponseWriter, r *http.Request) {
	if !c.Status().Healthy() {
		http.Error(w, "", http.StatusServiceUnavailable)
		return
	}
	// implicitly returns 200
}
