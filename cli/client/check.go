package client

import (
	"encoding/json"

	"github.com/sensu/sensu-go/types"
)

// CreateCheck creates new check on configured Sensu instance
func (client *RestClient) CreateCheck(check *types.CheckConfig) (err error) {
	bytes, err := json.Marshal(check)
	if err != nil {
		return err
	}

	res, err := client.R().SetBody(bytes).Post("/checks")
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return unmarshalError(res)
	}

	return nil
}

// UpdateCheck updates given check on configured Sensu instance
func (client *RestClient) UpdateCheck(check *types.CheckConfig) (err error) {
	bytes, err := json.Marshal(check)
	if err != nil {
		return err
	}

	res, err := client.R().SetBody(bytes).Patch("/checks/" + check.Name)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return unmarshalError(res)
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
		return unmarshalError(res)
	}

	return nil
}

// FetchCheck fetches a specific check
func (client *RestClient) FetchCheck(name string) (*types.CheckConfig, error) {
	var check *types.CheckConfig

	res, err := client.R().Get("/checks/" + name)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, unmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &check)
	return check, err
}

// ListChecks fetches all checks from configured Sensu instance
func (client *RestClient) ListChecks(org string) ([]types.CheckConfig, error) {
	var checks []types.CheckConfig
	res, err := client.R().SetQueryParam("org", org).Get("/checks")
	if err != nil {
		return checks, err
	}

	if res.StatusCode() >= 400 {
		return checks, unmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &checks)
	return checks, err
}
