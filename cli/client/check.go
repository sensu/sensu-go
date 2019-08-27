package client

import (
	"encoding/json"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// ChecksPath is the api path for checks.
var ChecksPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "checks")

// CreateCheck creates new check on configured Sensu instance
func (client *RestClient) CreateCheck(check *corev2.CheckConfig) (err error) {
	bytes, err := json.Marshal(check)
	if err != nil {
		return err
	}

	path := ChecksPath(check.Namespace)
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
func (client *RestClient) UpdateCheck(check *corev2.CheckConfig) (err error) {
	bytes, err := json.Marshal(check)
	if err != nil {
		return err
	}

	path := ChecksPath(check.Namespace, check.Name)
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
func (client *RestClient) DeleteCheck(namespace, name string) error {
	return client.Delete(ChecksPath(namespace, name))
}

// ExecuteCheck sends an execution request with the provided adhoc request
func (client *RestClient) ExecuteCheck(req *corev2.AdhocRequest) error {
	bytes, err := json.Marshal(req)
	if err != nil {
		return err
	}

	path := ChecksPath(client.config.Namespace(), req.Name, "execute")
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
func (client *RestClient) FetchCheck(name string) (*corev2.CheckConfig, error) {
	var check *corev2.CheckConfig

	path := ChecksPath(client.config.Namespace(), name)
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

// AddCheckHook associates an existing hook with an existing check
func (client *RestClient) AddCheckHook(check *corev2.CheckConfig, checkHook *corev2.HookList) error {
	path := ChecksPath(check.Namespace, check.Name, "hooks", checkHook.Type)
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
func (client *RestClient) RemoveCheckHook(check *corev2.CheckConfig, checkHookType string, hookName string) error {
	path := ChecksPath(check.Namespace, check.Name, "hooks", checkHookType, "hook", hookName)
	res, err := client.R().Delete(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}
