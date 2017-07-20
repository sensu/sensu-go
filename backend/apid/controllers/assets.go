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

// AssetsController defines those fields required.
type AssetsController struct {
	Store     store.Store
	abilities authorization.Ability
}

// Register defines an association between HTTP routes and their respective
// handlers defined within this controller.
func (c *AssetsController) Register(r *mux.Router) {
	c.abilities = authorization.Ability{Resource: types.RuleTypeAsset}

	r.HandleFunc("/assets", c.many).Methods(http.MethodGet)
	r.HandleFunc("/assets/{name}", c.single).Methods(http.MethodGet, http.MethodPut, http.MethodPost)
}

// many handles requests to /assets
func (c *AssetsController) many(w http.ResponseWriter, r *http.Request) {
	abilities := c.abilities.WithContext(r.Context())
	if r.Method == http.MethodGet && !abilities.CanRead() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	assets, err := c.Store.GetAssets(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	assetBytes, err := json.Marshal(assets)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", assetBytes)
}

// signle handles requests to /assets/{name}
func (c *AssetsController) single(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, _ := vars["name"]
	method := r.Method

	abilities := c.abilities.WithContext(r.Context())
	if r.Method == http.MethodGet && !abilities.CanRead() {
		authorization.UnauthorizedAccessToResource(w)
		return
	}

	var (
		asset *types.Asset
		err   error
	)

	if name != "" && method == http.MethodGet || method == http.MethodDelete {
		asset, err = c.Store.GetAssetByName(r.Context(), name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if asset == nil {
			http.NotFound(w, r)
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		assetBytes, err := json.Marshal(asset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, string(assetBytes))
	case http.MethodPut, http.MethodPost:
		switch {
		case asset == nil && !abilities.CanCreate():
			fallthrough
		case asset != nil && !abilities.CanUpdate():
			authorization.UnauthorizedAccessToResource(w)
			return
		}

		newAsset := &types.Asset{}
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		err = json.Unmarshal(bodyBytes, newAsset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = newAsset.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = c.Store.UpdateAsset(r.Context(), newAsset); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	}
}
