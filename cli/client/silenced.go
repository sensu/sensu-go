package client

import (
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

// CreateSilenced creates a new silenced  entry.
func (client *RestClient) CreateSilenced(silenced *types.Silenced) error {
	b, err := json.Marshal(silenced)
	if err != nil {
		return err
	}

	res, err := client.R().SetBody(b).Post("/silenced")
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return unmarshalError(res)
	}

	return nil
}

// DeleteSilenced deletes a silenced entry.
func (client *RestClient) DeleteSilenced(id string) error {
	res, err := client.R().Delete(fmt.Sprintf("/silenced/%s", id))
	if err != nil {
		return err
	}
	if res.StatusCode() >= 400 {
		return unmarshalError(res)
	}
	return nil
}
