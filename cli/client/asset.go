package client

import (
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

// ListAssets fetches a list of asset resources from the backend
func (client *RestClient) ListAssets(org string) ([]types.Asset, error) {
	var assets []types.Asset

	res, err := client.R().Get("/assets?org=" + org)
	if err != nil {
		return assets, err
	}

	if res.StatusCode() >= 400 {
		return assets, fmt.Errorf("%v", res.String())
	}

	err = json.Unmarshal(res.Body(), &assets)
	return assets, err
}

// CreateAsset fetches an asset resource from the backend
func (client *RestClient) CreateAsset(asset *types.Asset) error {
	bytes, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	res, err := client.R().
		SetBody(bytes).
		Put("/assets/" + asset.Name)

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}
