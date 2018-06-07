package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

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
		Post(fmt.Sprintf("/rbac/organizations/%s/environments", url.PathEscape(org)))

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
	org, env = url.PathEscape(org), url.PathEscape(env)
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
		fmt.Sprintf("/rbac/organizations/%s/environments", url.PathEscape(org)),
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

// FetchEnvironment fetches an environment by name
func (client *RestClient) FetchEnvironment(envName string) (*types.Environment, error) {
	var env *types.Environment
	path := fmt.Sprintf(
		"/rbac/organizations/%s/environments/%s",
		url.PathEscape(client.config.Organization()),
		url.PathEscape(envName),
	)

	res, err := client.R().Get(path)
	if err != nil {
		return env, err
	}

	if res.StatusCode() >= 400 {
		return env, fmt.Errorf("error getting environment: %v", res.String())
	}

	err = json.Unmarshal(res.Body(), &env)
	return env, err
}

// UpdateEnvironment updates an existing environment.
func (client *RestClient) UpdateEnvironment(env *types.Environment) error {
	b, err := json.Marshal(env)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/rbac/organizations/%s/environments/%s",
		url.PathEscape(env.Organization), url.PathEscape(env.Name))
	res, err := client.R().SetBody(b).Put(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return errors.New(res.String())
	}

	return nil
}
