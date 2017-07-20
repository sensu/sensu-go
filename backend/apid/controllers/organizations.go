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

// OrganizationsController defines the fields required for this controller.
type OrganizationsController struct {
	Store     store.Store
	abilities authorization.Ability
}

// Register should define an association between HTTP routes and their
// respective handlers defined within this Controller.
func (o *OrganizationsController) Register(r *mux.Router) {
	o.abilities = authorization.Ability{Resource: types.RuleTypeOrganization}

	r.HandleFunc("/rbac/organizations", o.many).Methods(http.MethodGet)
	r.HandleFunc("/rbac/organizations", o.update).Methods(http.MethodPost)
	r.HandleFunc("/rbac/organizations/{organization}", o.single).Methods(http.MethodGet)
	r.HandleFunc("/rbac/organizations/{organization}", o.delete).Methods(http.MethodDelete)
}

// delete handles deletion of a specific organization
func (o *OrganizationsController) delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	org := vars["organization"]

	abilities := o.abilities.WithContext(r.Context())
	abilities.Actor.Organization = org

	if !abilities.CanDelete() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	err := o.Store.DeleteOrganizationByName(r.Context(), org)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusAccepted)
	return
}

// many returns all organizations
func (o *OrganizationsController) many(w http.ResponseWriter, r *http.Request) {
	abilities := o.abilities.WithContext(r.Context())
	if !abilities.CanRead() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	orgs, err := o.Store.GetOrganizations(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(orgs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(bytes))
}

// single returns a specific organization
func (o *OrganizationsController) single(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["organization"]

	abilities := o.abilities.WithContext(r.Context())
	abilities.Actor.Organization = name

	if !abilities.CanRead() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	var (
		org *types.Organization
		err error
	)

	org, err = o.Store.GetOrganizationByName(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if org == nil {
		http.NotFound(w, r)
		return
	}

	bytes, err := json.Marshal(org)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(bytes))

}

// update handles the update of a specific organization
func (o *OrganizationsController) update(w http.ResponseWriter, r *http.Request) {
	var org types.Organization

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	abilities := o.abilities.WithContext(r.Context())
	abilities.Actor.Organization = org.Name

	if !abilities.CanCreate() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	err = json.Unmarshal(bodyBytes, &org)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = org.Validate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = o.Store.UpdateOrganization(r.Context(), &org)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	return
}
