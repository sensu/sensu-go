package client

import (
	"encoding/json"
	"path"

	"github.com/sensu/sensu-go/types"
)

const checksBasePath = "/checks"

func checksPath(ext ...string) string {
	components := append([]string{checksBasePath}, ext...)
	return path.Join(components...)
}

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

// AddCheckHook associates an existing hook with an existing check
func (client *RestClient) AddCheckHook(check *types.CheckConfig, checkHook *types.HookList) error {
	key := checksPath(check.Name, "hooks", checkHook.Type)
	res, err := client.R().SetQueryParam("org", check.Organization).SetQueryParam("env", check.Environment).SetBody(checkHook).Put(key)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return unmarshalError(res)
	}

	return nil
}

// RemoveCheckHook removes an association between an existing hook and an existing check
func (client *RestClient) RemoveCheckHook(check *types.CheckConfig, checkHookType string, hookName string) error {
	path := checksPath(check.Name, "hooks", checkHookType, "hook", hookName)
	res, err := client.R().SetQueryParam("org", check.Organization).SetQueryParam("env", check.Environment).Delete(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return unmarshalError(res)
	}

	return nil
}
