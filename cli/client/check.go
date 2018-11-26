package client

import (
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

var checksPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "checks")

// CreateCheck creates new check on configured Sensu instance
func (client *RestClient) CreateCheck(check *types.CheckConfig) (err error) {
	bytes, err := json.Marshal(check)
	if err != nil {
		return err
	}

	path := checksPath(check.Namespace)
	res, err := client.R().SetBody(bytes).Post(path)
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

	path := checksPath(check.Namespace, check.Name)
	res, err := client.R().SetBody(bytes).Put(path)
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
	path := checksPath(client.config.Namespace(), check.Name)
	res, err := client.R().Delete(path)

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

	path := checksPath(client.config.Namespace(), req.Name, "execute")
	res, err := client.R().SetBody(bytes).Post(path)

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

	path := checksPath(client.config.Namespace(), name)
	res, err := client.R().Get(path)
	if err != nil {
		return nil, fmt.Errorf("GET %q: %s", path, err)
	}

	if res.StatusCode() >= 400 {
		return nil, UnmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &check)
	return check, err
}

// ListChecks fetches all checks from configured Sensu instance
func (client *RestClient) ListChecks(namespace string) ([]types.CheckConfig, error) {
	var checks []types.CheckConfig

	path := checksPath(namespace)
	res, err := client.R().Get(path)
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
	path := checksPath(check.Namespace, check.Name, "hooks", checkHook.Type)
	res, err := client.R().SetBody(checkHook).Put(path)
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
	path := checksPath(check.Namespace, check.Name, "hooks", checkHookType, "hook", hookName)
	res, err := client.R().Delete(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}
