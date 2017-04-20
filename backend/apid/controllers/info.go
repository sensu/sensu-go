package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// InfoController defines the fields required by InfoController.
type InfoController struct {
	Store  store.Store
	Status func() types.StatusMap
}

// Register should define an association between HTTP routes and their
// respective handlers defined within this Controller.
func (c *InfoController) Register(r *mux.Router) {
	r.HandleFunc("/info", c.many).Methods(http.MethodGet)
}

// many handles requests to /info
func (c *InfoController) many(w http.ResponseWriter, r *http.Request) {
	sb, err := json.Marshal(c.Status())
	if err != nil {
		logger.Error("error marshaling status: ", err.Error())
		http.Error(w, "Error getting server status.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(sb))
}
