package client

import (
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

// CreateEnvironment creates new env on configured Sensu instance
func (client *RestClient) CreateEnvironment(org string, env *types.Environment) error {
	bytes, err := json.Marshal(env)
	if err != nil {
		return err
	}

	res, err := client.R().
		SetBody(bytes).
		Post(fmt.Sprintf("/rbac/organizations/%s/environments", org))

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// DeleteEnvironment deletes an environment on configured Sensu instance
func (client *RestClient) DeleteEnvironment(org, env string) error {
	res, err := client.R().Delete(
		fmt.Sprintf("/rbac/organizations/%s/environments/%s", org, env),
	)

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// ListEnvironments fetches all organizations from configured Sensu instance
func (client *RestClient) ListEnvironments(org string) ([]types.Environment, error) {
	var envs []types.Environment

	res, err := client.R().Get(
		fmt.Sprintf("/rbac/organizations/%s/environments", org),
	)
	if err != nil {
		return envs, err
	}

	if res.StatusCode() >= 400 {
		return envs, fmt.Errorf("%v", res.String())
	}

	err = json.Unmarshal(res.Body(), &envs)
	return envs, err
}
