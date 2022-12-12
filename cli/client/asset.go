package client

import (
	"encoding/json"
	"fmt"

	corev2 "github.com/sensu/core/v2"
)

// AssetsPath is the api path for assets.
var AssetsPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "assets")

// FetchAsset fetches an asset resource from the backend
func (client *RestClient) FetchAsset(name string) (*corev2.Asset, error) {
	var asset corev2.Asset

	path := AssetsPath(client.config.Namespace(), name)
	res, err := client.R().Get(path)
	if err != nil {
		return &asset, fmt.Errorf("GET %q: %s", path, err)
	}

	if res.StatusCode() >= 400 {
		return &asset, UnmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &asset)
	return &asset, err
}

// CreateAsset creates an asset resource from the backend
func (client *RestClient) CreateAsset(asset *corev2.Asset) error {
	bytes, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	path := AssetsPath(asset.Namespace)
	res, err := client.R().SetBody(bytes).Post(path)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// UpdateAsset updates an asset resource from the backend
func (client *RestClient) UpdateAsset(asset *corev2.Asset) (err error) {
	bytes, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	path := AssetsPath(asset.Namespace, asset.Name)
	res, err := client.R().SetBody(bytes).Put(path)
	if err != nil {
		return fmt.Errorf("PUT %q: %s", path, err)
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}
