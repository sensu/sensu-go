package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"

	"github.com/sensu/sensu-go/types"
)

const checksBasePath = "/checks"

func checksPath(ext ...string) string {
	parts := ext
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return path.Join(append([]string{checksBasePath}, parts...)...)
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
		return UnmarshalError(res)
	}

	return nil
}

// UpdateCheck updates given check on configured Sensu instance
func (client *RestClient) UpdateCheck(check *types.CheckConfig) (err error) {
	bytes, err := json.Marshal(check)
	if err != nil {
		return err
	}

	checkPath := fmt.Sprintf("/checks/%s", url.PathEscape(check.Name))
	res, err := client.R().SetBody(bytes).Put(checkPath)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// DeleteCheck deletes check from configured Sensu instance
func (client *RestClient) DeleteCheck(check *types.CheckConfig) error {
	res, err := client.R().Delete("/checks/" + url.PathEscape(check.Name))

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// ExecuteCheck sends an execution request with the provided adhoc request
func (client *RestClient) ExecuteCheck(req *types.AdhocRequest) error {
	bytes, err := json.Marshal(req)
	if err != nil {
		return err
	}

	checkPath := fmt.Sprintf("/checks/%s/execute", url.PathEscape(req.Name))
	res, err := client.R().
		SetQueryParam("env", client.config.Environment()).
		SetQueryParam("org", client.config.Organization()).
		SetBody(bytes).
		Post(checkPath)

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// FetchCheck fetches a specific check
func (client *RestClient) FetchCheck(name string) (*types.CheckConfig, error) {
	var check *types.CheckConfig

	checkPath := fmt.Sprintf("/checks/%s", url.PathEscape(name))
	res, err := client.R().Get(checkPath)
	if err != nil {
		return nil, fmt.Errorf("GET %q: %s", checkPath, err)
	}

	if res.StatusCode() >= 400 {
		return nil, UnmarshalError(res)
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
		return checks, UnmarshalError(res)
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
		return UnmarshalError(res)
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
		return UnmarshalError(res)
	}

	return nil
}
