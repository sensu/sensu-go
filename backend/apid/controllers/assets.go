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
	Store store.Store
}

// Register defines an association between HTTP routes and their respective
// handlers defined within this controller.
func (c *AssetsController) Register(r *mux.Router) {
	r.HandleFunc("/assets", c.many).Methods(http.MethodGet)
	r.NewRoute().
		Path("/assets/{name:"+types.AssetNameRegexStr+"}").
		Methods(http.MethodGet, http.MethodPut, http.MethodPost).
		HandlerFunc(c.single)
}

// many handles requests to /assets
func (c *AssetsController) many(w http.ResponseWriter, r *http.Request) {
	abilities := authorization.Assets.WithContext(r.Context())
	if r.Method == http.MethodGet && !abilities.CanList() {
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

	// Reject those resources the viewer is unauthorized to view
	rejectAssets(&assets, abilities.CanRead)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", assetBytes)
}

// signle handles requests to /assets/{name}
func (c *AssetsController) single(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, _ := vars["name"]
	method := r.Method

	abilities := authorization.Assets.WithContext(r.Context())

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

		if !abilities.CanRead(asset) {
			authorization.UnauthorizedAccessToResource(w)
			return
		}

		fmt.Fprint(w, string(assetBytes))
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

func rejectAssets(records *[]*types.Asset, predicate func(*types.Asset) bool) {
	for i := 0; i < len(*records); i++ {
		if !predicate((*records)[i]) {
			*records = append((*records)[:i], (*records)[i+1:]...)
			i--
		}
	}
}
