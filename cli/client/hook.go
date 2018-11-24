package client

import (
	"encoding/json"

	"github.com/sensu/sensu-go/types"
)

var hooksPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "hooks")

// CreateHook creates new hook on configured Sensu instance
func (client *RestClient) CreateHook(hook *types.HookConfig) (err error) {
	bytes, err := json.Marshal(hook)
	if err != nil {
		return err
	}

	path := hooksPath(hook.Namespace)
	res, err := client.R().SetBody(bytes).Post(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// UpdateHook updates given hook on configured Sensu instance
func (client *RestClient) UpdateHook(hook *types.HookConfig) (err error) {
	bytes, err := json.Marshal(hook)
	if err != nil {
		return err
	}

	path := hooksPath(hook.Namespace, hook.Name)
	res, err := client.R().SetBody(bytes).Put(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// DeleteHook deletes hook from configured Sensu instance
func (client *RestClient) DeleteHook(hook *types.HookConfig) error {
	path := hooksPath(hook.Namespace, hook.Name)
	res, err := client.R().Delete(path)

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// FetchHook fetches a specific hook
func (client *RestClient) FetchHook(name string) (*types.HookConfig, error) {
	var hook *types.HookConfig

	path := hooksPath(client.config.Namespace(), name)
	res, err := client.R().Get(path)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, UnmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &hook)
	return hook, err
}

// ListHooks fetches all hooks from configured Sensu instance
func (client *RestClient) ListHooks(namespace string) ([]types.HookConfig, error) {
	var hooks []types.HookConfig

	path := hooksPath(namespace)
	res, err := client.R().Get(path)
	if err != nil {
		return hooks, err
	}

	if res.StatusCode() >= 400 {
		return hooks, UnmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &hooks)
	return hooks, err
}
