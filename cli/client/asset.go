package client

import (
	"encoding/json"
	"fmt"
	"net/url"

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

// FetchAsset fetches an asset resource from the backend
func (client *RestClient) FetchAsset(name string) (*types.Asset, error) {
	var asset types.Asset

	assetPath := fmt.Sprintf("/assets/%s", url.PathEscape(name))
	res, err := client.R().Get(assetPath)
	if err != nil {
		return &asset, fmt.Errorf("GET %q: %s", assetPath, err)
	}

	if res.StatusCode() >= 400 {
		return &asset, fmt.Errorf("GET %q: %s", assetPath, res.String())
	}

	err = json.Unmarshal(res.Body(), &asset)
	return &asset, err
}

// CreateAsset creates an asset resource from the backend
func (client *RestClient) CreateAsset(asset *types.Asset) error {
	bytes, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	res, err := client.R().SetBody(bytes).Post("/assets")
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// UpdateAsset updates an asset resource from the backend
func (client *RestClient) UpdateAsset(asset *types.Asset) (err error) {
	bytes, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	assetPath := fmt.Sprintf("/assets/%s", url.PathEscape(asset.Name))
	res, err := client.R().SetBody(bytes).Put(assetPath)
	if err != nil {
		return fmt.Errorf("PUT %q: %s", assetPath, err)
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("PUT %q: %s", assetPath, res.String())
	}

	return nil
}
