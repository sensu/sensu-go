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

// EnvironmentsController defines the fields required for this controller.
type EnvironmentsController struct {
	Store store.Store
}

// Register should define an association between HTTP routes and their
// respective handlers defined within this Controller.
func (o *EnvironmentsController) Register(r *mux.Router) {
	r.HandleFunc("/rbac/organizations/{organization}/environments", o.many).Methods(http.MethodGet)
	r.HandleFunc("/rbac/organizations/{organization}/environments", o.update).Methods(http.MethodPost)
	r.HandleFunc("/rbac/organizations/{organization}/environments/{environment}", o.single).Methods(http.MethodGet)
	r.HandleFunc("/rbac/organizations/{organization}/environments/{environment}", o.delete).Methods(http.MethodDelete)
}

// delete handles deletion of a specific environment
func (o *EnvironmentsController) delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	env := &types.Environment{
		Name:         vars["environment"],
		Organization: vars["organization"],
	}

	if err := env.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := o.Store.DeleteEnvironment(r.Context(), env)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	return
}

// many returns all environments within an organization
func (o *EnvironmentsController) many(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	org := vars["organization"]

	envs, err := o.Store.GetEnvironments(r.Context(), org)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(envs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(bytes))
}

// single returns a specific environment
func (o *EnvironmentsController) single(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	org := vars["organization"]
	env := vars["environment"]

	environment, err := o.Store.GetEnvironment(r.Context(), org, env)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if environment == nil {
		http.NotFound(w, r)
		return
	}

	bytes, err := json.Marshal(environment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(bytes))

}

// update handles the update of a specific environment
func (o *EnvironmentsController) update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	org := vars["organization"]

	env := types.Environment{
		Organization: org,
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(bodyBytes, &env)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = env.Validate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = o.Store.UpdateEnvironment(r.Context(), &env)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	return
}
