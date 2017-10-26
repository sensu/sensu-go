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

// SilencedController defines the fields required by SilencedController.
type SilencedController struct {
	Store store.Store
}

// Register should define an association between HTTP routes and their
// respective handlers defined within this Controller.
func (c *SilencedController) Register(r *mux.Router) {
	r.HandleFunc("/silenced", c.all).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/silenced/{id}", c.byID).Methods(http.MethodGet, http.MethodDelete)
	r.HandleFunc("/silenced/subscriptions/{subscription}", c.bySubscription).Methods(http.MethodGet, http.MethodDelete)
	r.HandleFunc("/silenced/checks/{check}", c.byCheck).Methods(http.MethodGet, http.MethodDelete)
}

// all handles GET requests to /silenced
func (c *SilencedController) all(w http.ResponseWriter, r *http.Request) {
	abilities := authorization.Silenced.WithContext(r.Context())

	switch r.Method {
	case http.MethodGet:
		if !abilities.CanList() {
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
	case http.MethodPost:
		silencedEntry := &types.Silenced{}
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = json.Unmarshal(bodyBytes, silencedEntry)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		if !abilities.CanCreate(silencedEntry) {
			authorization.UnauthorizedAccessToResource(w)
			return
		}
		// Populate silencedEntry.ID with the subscription and checkName. Substitute a
		// splat if one of the values does not exist. If both values are empty, the
		// validator will return an error when attempting to update it in the store.
		if silencedEntry.Subscription != "" && silencedEntry.CheckName != "" {
			silencedEntry.ID = silencedEntry.Subscription + ":" + silencedEntry.CheckName
		} else if silencedEntry.CheckName == "" && silencedEntry.Subscription != "" {
			silencedEntry.ID = silencedEntry.Subscription + ":" + "*"
		} else if silencedEntry.Subscription == "" && silencedEntry.CheckName != "" {
			silencedEntry.ID = "*" + ":" + silencedEntry.CheckName
		}

		err = c.Store.UpdateSilencedEntry(r.Context(), silencedEntry)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

// byID handles requests to /silenced/:id
func (c *SilencedController) byID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	silencedID, _ := vars["id"]
	abilities := authorization.Silenced.WithContext(r.Context())
	switch r.Method {
	case http.MethodGet:
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
	case http.MethodDelete:
		if !abilities.CanDelete() {
			authorization.UnauthorizedAccessToResource(w)
			return
		}

		err := c.Store.DeleteSilencedEntryByID(r.Context(), silencedID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}
}

func (c *SilencedController) bySubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subscription, _ := vars["subscription"]
	abilities := authorization.Silenced.WithContext(r.Context())
	switch r.Method {
	case http.MethodGet:
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

	case http.MethodDelete:
		if !abilities.CanDelete() {
			authorization.UnauthorizedAccessToResource(w)
			return
		}

		err := c.Store.DeleteSilencedEntryBySubscription(r.Context(), subscription)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}
}

func (c *SilencedController) byCheck(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	checkName, _ := vars["check"]
	abilities := authorization.Silenced.WithContext(r.Context())

	switch r.Method {
	case http.MethodGet:
		if !abilities.CanList() {
			authorization.UnauthorizedAccessToResource(w)
			return
		}

		var (
			silencedEntries []*types.Silenced
			err             error
		)
		silencedEntries, err = c.Store.GetSilencedEntriesByCheckName(r.Context(), checkName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// Reject those resources the viewer is unauthorized to view
		rejectSilencedEntries(&silencedEntries, abilities.CanRead)

		silencedBytes, err := json.Marshal(silencedEntries)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, string(silencedBytes))
	case http.MethodDelete:
		if !abilities.CanDelete() {
			authorization.UnauthorizedAccessToResource(w)
			return
		}

		err := c.Store.DeleteSilencedEntryByCheckName(r.Context(), checkName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}
}

func rejectSilencedEntries(records *[]*types.Silenced, predicate func(*types.Silenced) bool) {
	for i := 0; i < len(*records); i++ {
		if !predicate((*records)[i]) {
			*records = append((*records)[:i], (*records)[i+1:]...)
			i--
		}
	}
}
