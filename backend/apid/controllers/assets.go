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

// AssetsController defines those fields required.
type AssetsController struct {
	Store store.Store
}

// Register defines an association between HTTP routes and their respective
// handlers defined within this controller.
func (c *AssetsController) Register(r *mux.Router) {
	r.HandleFunc("/assets", c.many).Methods(http.MethodGet)
	r.HandleFunc("/assets/{name}", c.single).Methods(http.MethodGet, http.MethodPut, http.MethodPost)
}

// many handles requests to /assets
func (c *AssetsController) many(w http.ResponseWriter, r *http.Request) {
	assets, err := c.Store.GetAssets()
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

	var (
		asset *types.Asset
		err   error
	)

	if method == http.MethodGet || method == http.MethodDelete {
		asset, err = c.Store.GetAssetByName(name)
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

		if len(newAsset.Hash) == 0 {
			if err := newAsset.UpdateHash(); err != nil {
				http.Error(w, "unable to read given URL", http.StatusBadRequest)
				return
			}
		}

		if err = newAsset.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = c.Store.UpdateAsset(newAsset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
}
