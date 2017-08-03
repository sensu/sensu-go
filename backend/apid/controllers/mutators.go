package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// MutatorsController defines the fields required by MutatorsController.
type MutatorsController struct {
	Store store.Store
}

// Register should define an association between HTTP routes and their
// respective handlers defined within this Controller.
func (c *MutatorsController) Register(r *mux.Router) {
	r.HandleFunc("/mutators", c.many).Methods(http.MethodGet)
	r.HandleFunc("/mutators/{name}", c.single).Methods(http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete)
}

// many handles requests to /mutators
func (c *MutatorsController) many(w http.ResponseWriter, r *http.Request) {
	abilities := authorization.Mutators.WithContext(r.Context())
	if r.Method == http.MethodGet && !abilities.CanList() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	mutators, err := c.Store.GetMutators(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reject those resources the viewer is unauthorized to view
	rejectMutators(&mutators, abilities.CanRead)

	mutatorsBytes, err := json.Marshal(mutators)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(mutatorsBytes))
}

// single handles requests to /mutators/:name
func (c *MutatorsController) single(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	method := r.Method

	abilities := authorization.Mutators.WithContext(r.Context())
	if method == http.MethodDelete && !abilities.CanDelete() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	var (
		mutator *types.Mutator
		err     error
	)

	if method == http.MethodGet || method == http.MethodDelete {
		mutator, err = c.Store.GetMutatorByName(r.Context(), name)
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
		if !abilities.CanRead(mutator) {
			authorization.UnauthorizedAccessToResource(w)
			return
		}

		mutatorBytes, err := json.Marshal(mutator)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(mutatorBytes))
	case http.MethodPut, http.MethodPost:
		switch {
		case mutator == nil && !abilities.CanCreate():
			fallthrough
		case mutator != nil && !abilities.CanUpdate():
			authorization.UnauthorizedAccessToResource(w)
			return
		}

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

		err = c.Store.UpdateMutator(r.Context(), newMutator)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodDelete:
		err := c.Store.DeleteMutatorByName(r.Context(), name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func rejectMutators(records *[]*types.Mutator, predicate func(*types.Mutator) bool) {
	for i := 0; i < len(*records); i++ {
		if !predicate((*records)[i]) {
			*records = append((*records)[:i], (*records)[i+1:]...)
			i--
		}
	}
}
