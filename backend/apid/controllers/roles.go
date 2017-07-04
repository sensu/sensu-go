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
	roles, err := c.Store.GetRoles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	switch method {
	case http.MethodGet:
		roleBytes, err := json.Marshal(role)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(roleBytes))
	case http.MethodPut, http.MethodPost:
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
