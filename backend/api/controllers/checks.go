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

type ChecksController struct {
	Store store.Store
}

func (c *ChecksController) Register(r *mux.Router) {
	r.HandleFunc("/checks", c.many).Methods(http.MethodGet)
	r.HandleFunc("/checks/{name}", c.single).Methods(http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete)
}

// many handles requests to /checks
func (a *ChecksController) many(w http.ResponseWriter, r *http.Request) {
	checks, err := a.Store.GetChecks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	checksBytes, err := json.Marshal(checks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Fprintf(w, string(checksBytes))
}

// single handles requests to /checks/:name
func (a *ChecksController) single(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	method := r.Method

	var (
		check *types.Check
		err   error
	)

	if method == http.MethodGet || method == http.MethodDelete {
		check, err = a.Store.GetCheckByName(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if check == nil {
			http.NotFound(w, r)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		checkBytes, err := json.Marshal(check)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, string(checkBytes))
	case http.MethodPut, http.MethodPost:
		newCheck := &types.Check{}
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		err = json.Unmarshal(bodyBytes, newCheck)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = newCheck.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = a.Store.UpdateCheck(newCheck)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	case http.MethodDelete:
		err := a.Store.DeleteCheckByName(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
}
