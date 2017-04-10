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

type HandlersController struct {
	Store store.Store
}

func (c *HandlersController) Register(r *mux.Router) {
	r.HandleFunc("/handlers", c.many).Methods(http.MethodGet)
	r.HandleFunc("/handlers/{name}", c.single).Methods(http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete)
}

// many handles requests to /handlers
func (a *HandlersController) many(w http.ResponseWriter, r *http.Request) {
	handlers, err := a.Store.GetHandlers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	handlersBytes, err := json.Marshal(handlers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Fprintf(w, string(handlersBytes))
}

// single handles requests to /handlers/:name
func (a *HandlersController) single(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	method := r.Method

	var (
		handler *types.Handler
		err     error
	)

	if method == http.MethodGet || method == http.MethodDelete {
		handler, err = a.Store.GetHandlerByName(name)
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

		fmt.Fprintf(w, string(handlerBytes))
	case http.MethodPut, http.MethodPost:
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

		err = a.Store.UpdateHandler(newHandler)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodDelete:
		err := a.Store.DeleteHandlerByName(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
