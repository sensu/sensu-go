package bonsai

import (
	"encoding/json"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// FetchAsset fetches an asset (list of versions)
func (client *RestClient) FetchAsset(namespace, name string) (*corev2.BonsaiAsset, error) {
	path := fmt.Sprintf("/%s/%s", namespace, name)
	res, err := client.R().Get(path)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		// TODO: this should return an actual error
		return nil, err
	}

	var asset corev2.BonsaiAsset
	if err = json.Unmarshal(res.Body(), &asset); err != nil {
		return nil, err
	}

	return &asset, nil
}
