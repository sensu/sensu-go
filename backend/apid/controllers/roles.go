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

// RolesController defines the fields required by RolesController.
type RolesController struct {
	Store store.Store
}

// Register should define an association between HTTP routes and their
// respective roles defined within this Controller.
func (c *RolesController) Register(r *mux.Router) {
	r.HandleFunc("/rbac/roles", c.many).Methods(http.MethodGet)
	r.HandleFunc("/rbac/roles/{name}", c.single).Methods(http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete)
	r.HandleFunc("/rbac/roles/{name}/rules/{type}", c.rules).Methods(http.MethodPut, http.MethodDelete)
}

func (c *RolesController) many(w http.ResponseWriter, r *http.Request) {
	abilities := authorization.Roles.WithContext(r.Context())
	if !abilities.CanList() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	roles, err := c.Store.GetRoles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reject those resources the viewer is unauthorized to view
	rejectRoles(&roles, abilities.CanRead)

	rolesBytes, err := json.Marshal(roles)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(rolesBytes))
}

func (c *RolesController) single(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	method := r.Method

	var (
		role *types.Role
		err  error
	)

	if method == http.MethodGet || method == http.MethodDelete {
		role, err = c.Store.GetRoleByName(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if role == nil {
			http.NotFound(w, r)
			return
		}
	}

	abilities := authorization.Roles.WithContext(r.Context())

	switch method {
	case http.MethodGet:
		if !abilities.CanRead(role) {
			authorization.UnauthorizedAccessToResource(w)
			return
		}

		roleBytes, err := json.Marshal(role)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(roleBytes))
	case http.MethodPut, http.MethodPost:
		if !abilities.CanCreate() {
			authorization.UnauthorizedAccessToResource(w)
			return
		}

		newRole := &types.Role{}
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		err = json.Unmarshal(bodyBytes, newRole)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = newRole.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = c.Store.UpdateRole(newRole)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodDelete:
		if !abilities.CanDelete() {
			authorization.UnauthorizedAccessToResource(w)
			return
		}

		err := c.Store.DeleteRoleByName(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (c *RolesController) rules(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	abilities := authorization.Roles.WithContext(r.Context())
	if !abilities.CanUpdate() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	role, err := c.Store.GetRoleByName(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if role == nil {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodPut:
		newRule := types.Rule{}
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		err = json.Unmarshal(bodyBytes, &newRule)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		role.Rules = append(role.Rules, newRule)
		if err = role.Validate(); err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	case http.MethodDelete:
		newRuleSet := []types.Rule{}
		ruleType := vars["type"]

		for _, rule := range role.Rules {
			if rule.Type != ruleType {
				newRuleSet = append(newRuleSet, rule)
			}
		}

		role.Rules = newRuleSet
	}

	err = c.Store.UpdateRole(role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func rejectRoles(records *[]*types.Role, predicate func(*types.Role) bool) {
	new := make([]*types.Role, 0, len(*records))
	for _, record := range *records {
		if predicate(record) {
			new = append(new, record)
		}
	}
	*records = new
}
