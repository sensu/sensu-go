package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// SilencedController defines the fields required by SilencedController.
type SilencedController struct {
	Store store.Store
}

// Register should define an association between HTTP routes and their
// respective handlers defined within this Controller.
func (c *SilencedController) Register(r *mux.Router) {
	r.HandleFunc("/silenced", c.many).Methods(http.MethodGet)
	r.HandleFunc("/silenced", c.update).Methods(http.MethodPost)
	r.HandleFunc("/silenced/ids/{id}", c.getByID).Methods(http.MethodGet)
	r.HandleFunc("/silenced/clear", c.clear).Methods(http.MethodPost)
	r.HandleFunc("/silenced/subscriptions/{subscription}", c.getBySubscription).Methods(http.MethodGet)
	r.HandleFunc("/silenced/checks/{check}", c.getByCheck).Methods(http.MethodGet)
}

// many handles requests to /silenced
func (c *SilencedController) many(w http.ResponseWriter, r *http.Request) {
	abilities := authorization.Silenced.WithContext(r.Context())
	if r.Method == http.MethodGet && !abilities.CanList() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	silencedEntries, err := c.Store.GetSilencedEntries(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reject those resources the viewer is unauthorized to view
	rejectSilencedEntries(&silencedEntries, abilities.CanRead)

	silencedBytes, err := json.Marshal(silencedEntries)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(silencedBytes))
}

// update handles POST resquests to /silenced
func (c *SilencedController) update(w http.ResponseWriter, r *http.Request) {
}

// getByID handles requests to /silenced/ids/:id
func (c *SilencedController) getByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	silencedID, _ := vars["id"]

	abilities := authorization.Silenced.WithContext(r.Context())
	if !abilities.CanList() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	var (
		silencedEntries []*types.Silenced
		err             error
	)

	silencedEntries, err = c.Store.GetSilencedEntryByID(r.Context(), silencedID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reject those resources the viewer is unauthorized to view
	rejectSilencedEntries(&silencedEntries, abilities.CanRead)

	silencedBytes, err := json.Marshal(silencedEntries)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(silencedBytes))
}

//
func (c *SilencedController) clear(w http.ResponseWriter, r *http.Request) {
}

func (c *SilencedController) getBySubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subscription, _ := vars["subscription"]

	abilities := authorization.Silenced.WithContext(r.Context())
	if !abilities.CanList() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	var (
		silencedEntries []*types.Silenced
		err             error
	)
	silencedEntries, err = c.Store.GetSilencedEntriesBySubscription(r.Context(), subscription)

	// Reject those resources the viewer is unauthorized to view
	rejectSilencedEntries(&silencedEntries, abilities.CanRead)

	silencedBytes, err := json.Marshal(silencedEntries)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(silencedBytes))
}

func (c *SilencedController) getByCheck(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	check, _ := vars["check"]

	abilities := authorization.Silenced.WithContext(r.Context())
	if !abilities.CanList() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	var (
		silencedEntries []*types.Silenced
		err             error
	)
	silencedEntries, err = c.Store.GetSilencedEntriesByCheckName(r.Context(), check)

	// Reject those resources the viewer is unauthorized to view
	rejectSilencedEntries(&silencedEntries, abilities.CanRead)

	silencedBytes, err := json.Marshal(silencedEntries)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(silencedBytes))
}

func rejectSilencedEntries(records *[]*types.Silenced, predicate func(*types.Silenced) bool) {
	for i := 0; i < len(*records); i++ {
		if !predicate((*records)[i]) {
			*records = append((*records)[:i], (*records)[i+1:]...)
			i--
		}
	}
}
