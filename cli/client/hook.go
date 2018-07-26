package client

import (
	"encoding/json"
	"net/url"

	"github.com/sensu/sensu-go/types"
)

// CreateHook creates new hook on configured Sensu instance
func (client *RestClient) CreateHook(hook *types.HookConfig) (err error) {
	bytes, err := json.Marshal(hook)
	if err != nil {
		return err
	}

	res, err := client.R().SetBody(bytes).Post("/hooks")
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

	res, err := client.R().SetBody(bytes).Put("/hooks/" + url.PathEscape(hook.Name))
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
	res, err := client.R().Delete("/hooks/" + url.PathEscape(hook.Name))

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

	res, err := client.R().Get("/hooks/" + url.PathEscape(name))
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
func (client *RestClient) ListHooks(org string) ([]types.HookConfig, error) {
	var hooks []types.HookConfig
	res, err := client.R().SetQueryParam("org", org).Get("/hooks")
	if err != nil {
		return hooks, err
	}

	if res.StatusCode() >= 400 {
		return hooks, UnmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &hooks)
	return hooks, err
}
