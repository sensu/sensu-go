package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

type MutatorsController struct {
	Store store.Store
}

func (c *MutatorsController) Register(r *mux.Router) {
	r.HandleFunc("/mutators", c.many).Methods(http.MethodGet)
	r.HandleFunc("/mutators/{name}", c.single).Methods(http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete)
}

// many handles requests to /mutators
func (a *MutatorsController) many(w http.ResponseWriter, r *http.Request) {
	mutators, err := a.Store.GetMutators()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mutatorsBytes, err := json.Marshal(mutators)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Fprintf(w, string(mutatorsBytes))
}

// single handles requests to /mutators/:name
func (a *MutatorsController) single(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	method := r.Method

	var (
		mutator *types.Mutator
		err     error
	)

	if method == http.MethodGet || method == http.MethodDelete {
		mutator, err = a.Store.GetMutatorByName(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if mutator == nil {
			http.NotFound(w, r)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		mutatorBytes, err := json.Marshal(mutator)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, string(mutatorBytes))
	case http.MethodPut, http.MethodPost:
		newMutator := &types.Mutator{}
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		err = json.Unmarshal(bodyBytes, newMutator)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = newMutator.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = a.Store.UpdateMutator(newMutator)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodDelete:
		err := a.Store.DeleteMutatorByName(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
