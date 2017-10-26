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

// FiltersController handles requests to /filters endpoint
type FiltersController struct {
	Store store.Store
}

// Register defines an association between HTTP routes and their
// respective handlers defined within this Controller.
func (c *FiltersController) Register(r *mux.Router) {
	r.HandleFunc("/filters", c.getFilters).Methods(http.MethodGet)
	r.HandleFunc("/filters/{name}", c.deleteFilter).Methods(http.MethodDelete)
	r.HandleFunc("/filters/{name}", c.getFilter).Methods(http.MethodGet)
	r.HandleFunc("/filters/{name}", c.updateFilter).Methods(http.MethodPost, http.MethodPut)
}

// deleteFilter deletes a specific filter, using its name
func (c *FiltersController) deleteFilter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	abilities := authorization.Filters.WithContext(r.Context())
	if !abilities.CanDelete() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	err := c.Store.DeleteEventFilterByName(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}

// getFilter returns a specific filter, using its name
func (c *FiltersController) getFilter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	filter, err := c.Store.GetEventFilterByName(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if filter == nil {
		http.NotFound(w, r)
		return
	}

	abilities := authorization.Filters.WithContext(r.Context())
	if !abilities.CanRead(filter) {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	filtersBytes, err := json.Marshal(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(filtersBytes))
}

// getFilters returns all available filters
func (c *FiltersController) getFilters(w http.ResponseWriter, r *http.Request) {
	abilities := authorization.Filters.WithContext(r.Context())
	if r.Method == http.MethodGet && !abilities.CanList() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	filters, err := c.Store.GetEventFilters(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reject those resources the viewer is unauthorized to view
	rejectFilters(&filters, abilities.CanRead)

	filtersBytes, err := json.Marshal(filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(filtersBytes))
}

// updateFilter creates or updates a filter
func (c *FiltersController) updateFilter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	filter := &types.EventFilter{}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(bodyBytes, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if filter.Name != name {
		http.Error(w,
			"The supplied filter name in the Request-URI does not match the filter data provided",
			http.StatusBadRequest,
		)
		return
	}

	if err = filter.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// The update permission is ignored so don't worry for now.
	// See https://github.com/sensu/sensu-go/issues/485
	abilities := authorization.Filters.WithContext(r.Context())
	if !abilities.CanCreate(filter) {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	err = c.Store.UpdateEventFilter(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func rejectFilters(records *[]*types.EventFilter, predicate func(*types.EventFilter) bool) {
	for i := 0; i < len(*records); i++ {
		if !predicate((*records)[i]) {
			*records = append((*records)[:i], (*records)[i+1:]...)
			i--
		}
	}
}
