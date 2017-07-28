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
	Store     store.Store
	abilities authorization.Ability
}

// Register should define an association between HTTP routes and their
// respective handlers defined within this Controller.
func (c *EntitiesController) Register(r *mux.Router) {
	c.abilities = authorization.Ability{Resource: types.RuleTypeEntity}

	r.HandleFunc("/entities", c.many).Methods(http.MethodGet)
	r.HandleFunc("/entities/{id}", c.single).Methods(http.MethodGet, http.MethodDelete)
}

func (c *EntitiesController) many(w http.ResponseWriter, r *http.Request) {
	policy := c.abilities.WithContext(r.Context())

	switch {
	case r.Method == http.MethodGet && !policy.CanRead():
		fallthrough
	case r.Method == http.MethodPost && !policy.CanCreate():
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	es, err := c.Store.GetEntities(r.Context())
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

	policy := c.abilities.WithContext(r.Context())
	switch {
	case r.Method == http.MethodGet && !policy.CanRead():
		fallthrough
	case r.Method == http.MethodDelete && !policy.CanDelete():
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
