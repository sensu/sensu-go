package client

import (
	"encoding/json"

	"github.com/sensu/sensu-go/types"
)

// ListAssets fetches a list of asset resources from the backend
func (client *RestClient) ListAssets() (assets []types.Asset, err error) {
	r, err := client.R().Get("/assets")
	if err == nil {
		err = json.Unmarshal(r.Body(), &assets)
	}

	return
}

// CreateAsset fetches an asset resource from the backend
func (client *RestClient) CreateAsset(asset *types.Asset) (err error) {
	bytes, err := json.Marshal(asset)
	if err == nil {
		_, err = client.R().
			SetBody(bytes).
			Put("/assets/" + asset.Name)
	}

	return
}
