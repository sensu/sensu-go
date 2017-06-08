package client

import (
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

// ListChecks fetches all checks from configured Sensu instance
func (client *RestClient) ListChecks() ([]types.CheckConfig, error) {
	var checks []types.CheckConfig
	res, err := client.R().Get("/checks")
	if err != nil {
		return checks, err
	}

	if res.StatusCode() >= 400 {
		return checks, fmt.Errorf("%v", res.String())
	}

	err = json.Unmarshal(res.Body(), &checks)
	return checks, err
}

// CreateCheck creates new check on configured Sensu instance
func (client *RestClient) CreateCheck(check *types.CheckConfig) (err error) {
	bytes, err := json.Marshal(check)
	if err != nil {
		return err
	}

	res, err := client.R().
		SetBody(bytes).
		Put("/checks/" + check.Name)

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// DeleteCheck deletes check from configured Sensu instance
func (client *RestClient) DeleteCheck(check *types.CheckConfig) error {
	res, err := client.R().Delete("/checks/" + check.Name)

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}
