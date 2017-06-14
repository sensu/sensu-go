package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// EntitiesController defines the fields required by EntitiesController.
type EntitiesController struct {
	Store store.Store
}

// Register should define an association between HTTP routes and their
// respective handlers defined within this Controller.
func (c *EntitiesController) Register(r *mux.Router) {
	r.HandleFunc("/entities", c.many).Methods(http.MethodGet)
	r.HandleFunc("/entities/{id}", c.single).Methods(http.MethodGet, http.MethodDelete)
}

// many handles GET requests to the /entities endpoint.
func (c *EntitiesController) many(w http.ResponseWriter, r *http.Request) {
	org := organization(r)

	es, err := c.Store.GetEntities(org)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	esb, err := json.Marshal(es)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(esb))
}

// single handles requests to /entities/{id}.
func (c *EntitiesController) single(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	method := r.Method

	w.Header().Set("Content-Type", "application/json")

	var (
		entity *types.Entity
		err    error
	)

	org := organization(r)

	if method == http.MethodGet || method == http.MethodDelete {
		entity, err = c.Store.GetEntityByID(org, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if entity == nil {
			http.NotFound(w, r)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		eb, err := json.Marshal(entity)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, string(eb))
	case http.MethodDelete:
		err := c.Store.DeleteEntityByID(org, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
