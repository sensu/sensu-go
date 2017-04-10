package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/store"
)

type EntitiesController struct {
	Store store.Store
}

func (c *EntitiesController) Register(r *mux.Router) {
	r.HandleFunc("/entities", c.many).Methods(http.MethodGet)
	r.HandleFunc("/entities/{name}", c.single).Methods(http.MethodGet)
}

// many handles GET requests to the /entities endpoint.
func (a *EntitiesController) many(w http.ResponseWriter, r *http.Request) {
	es, err := a.Store.GetEntities()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	esb, err := json.Marshal(es)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(esb))
}

// single handles requests to /entities/{id}.
func (a *EntitiesController) single(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	entity, err := a.Store.GetEntityByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entity == nil {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	eb, err := json.Marshal(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(eb))
}
