package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/authorization"
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

func (c *EntitiesController) many(w http.ResponseWriter, r *http.Request) {
	abilities := authorization.Entities.WithContext(r.Context())

	switch {
	case r.Method == http.MethodGet && !abilities.CanList():
		fallthrough
	case r.Method == http.MethodPost && !abilities.CanCreate():
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	es, err := c.Store.GetEntities(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reject those resources the viewer is unauthorized to view
	rejectEntities(&es, abilities.CanRead)

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

	policy := authorization.Entities.WithContext(r.Context())
	if r.Method == http.MethodDelete && !policy.CanDelete() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	var (
		entity *types.Entity
		err    error
	)

	if method == http.MethodGet || method == http.MethodDelete {
		entity, err = c.Store.GetEntityByID(r.Context(), id)
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
		if !policy.CanRead(entity) {
			authorization.UnauthorizedAccessToResource(w)
			return
		}

		eb, err := json.Marshal(entity)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, string(eb))
	case http.MethodDelete:
		err := c.Store.DeleteEntityByID(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func rejectEntities(records *[]*types.Entity, predicate func(*types.Entity) bool) {
	new := make([]*types.Entity, 0, len(*records))
	for _, record := range *records {
		if predicate(record) {
			new = append(new, record)
		}
	}
	*records = new
}
