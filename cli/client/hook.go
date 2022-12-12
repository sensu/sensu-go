package client

import (
	"encoding/json"

	corev2 "github.com/sensu/core/v2"
)

// HooksPath is the api path for hooks.
var HooksPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "hooks")

// CreateHook creates new hook on configured Sensu instance
func (client *RestClient) CreateHook(hook *corev2.HookConfig) (err error) {
	bytes, err := json.Marshal(hook)
	if err != nil {
		return err
	}

	path := HooksPath(hook.Namespace)
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
func (client *RestClient) UpdateHook(hook *corev2.HookConfig) (err error) {
	bytes, err := json.Marshal(hook)
	if err != nil {
		return err
	}

	path := HooksPath(hook.Namespace, hook.Name)
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
func (client *RestClient) DeleteHook(namespace, name string) error {
	return client.Delete(HooksPath(namespace, name))
}

// FetchHook fetches a specific hook
func (client *RestClient) FetchHook(name string) (*corev2.HookConfig, error) {
	var hook *corev2.HookConfig

	path := HooksPath(client.config.Namespace(), name)
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
