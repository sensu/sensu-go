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

// HandlersController defines the fields required by HandlersController.
type HandlersController struct {
	Store     store.Store
	abilities authorization.Ability
}

// Register should define an association between HTTP routes and their
// respective handlers defined within this Controller.
func (c *HandlersController) Register(r *mux.Router) {
	c.abilities = authorization.Ability{Resource: types.RuleTypeCheck}

	r.HandleFunc("/handlers", c.many).Methods(http.MethodGet)
	r.HandleFunc("/handlers/{name}", c.single).Methods(http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete)
}

// many handles requests to /handlers
func (c *HandlersController) many(w http.ResponseWriter, r *http.Request) {
	abilities := c.abilities.WithContext(r.Context())
	if r.Method == http.MethodGet && !abilities.CanRead() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	handlers, err := c.Store.GetHandlers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	handlersBytes, err := json.Marshal(handlers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(handlersBytes))
}

// single handles requests to /handlers/:name
func (c *HandlersController) single(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	method := r.Method

	abilities := c.abilities.WithContext(r.Context())
	switch {
	case r.Method == http.MethodGet && !abilities.CanRead():
		fallthrough
	case r.Method == http.MethodDelete && !abilities.CanDelete():
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	var (
		handler *types.Handler
		err     error
	)

	if method == http.MethodGet || method == http.MethodDelete {
		handler, err = c.Store.GetHandlerByName(r.Context(), name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if handler == nil {
			http.NotFound(w, r)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		handlerBytes, err := json.Marshal(handler)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(handlerBytes))
	case http.MethodPut, http.MethodPost:
		switch {
		case handler == nil && !abilities.CanCreate():
			fallthrough
		case handler != nil && !abilities.CanUpdate():
			authorization.UnauthorizedAccessToResource(w)
			return
		}

		newHandler := &types.Handler{}
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		err = json.Unmarshal(bodyBytes, newHandler)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = newHandler.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = c.Store.UpdateHandler(r.Context(), newHandler)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodDelete:
		err := c.Store.DeleteHandlerByName(r.Context(), name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
