package client

import (
	"encoding/json"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
)

var assetsPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "assets")

// ListAssets fetches a list of asset resources from the backend
func (client *RestClient) ListAssets(namespace string, options ListOptions) ([]corev2.Asset, string, error) {
	var assets []corev2.Asset
	var header string

	path := assetsPath(namespace)
	request := client.R()

	ApplyListOptions(request, options)

	res, err := request.Get(path)
	if err != nil {
		return assets, header, err
	}

	header = EntityLimitHeader(res.Header())

	if res.StatusCode() >= 400 {
		return assets, header, UnmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &assets)
	return assets, header, err
}

// FetchAsset fetches an asset resource from the backend
func (client *RestClient) FetchAsset(name string) (*types.Asset, error) {
	var asset types.Asset

	path := assetsPath(client.config.Namespace(), name)
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
func (client *RestClient) CreateAsset(asset *types.Asset) error {
	bytes, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	path := assetsPath(asset.Namespace)
	res, err := client.R().SetBody(bytes).Post(path)
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

	path := assetsPath(asset.Namespace, asset.Name)
	res, err := client.R().SetBody(bytes).Put(path)
	if err != nil {
		return fmt.Errorf("PUT %q: %s", path, err)
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("PUT %q: %s", path, res.String())
	}

	return nil
}
