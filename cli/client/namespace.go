package client

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/sensu/sensu-go/types"
)

// CreateNamespace creates new namespace on configured Sensu instance
func (client *RestClient) CreateNamespace(namespace *types.Namespace) error {
	bytes, err := json.Marshal(namespace)
	if err != nil {
		return err
	}

	res, err := client.R().
		SetBody(bytes).
		Post("/rbac/namespaces")

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// UpdateNamespace updates given namespace on a configured Sensu instance
func (client *RestClient) UpdateNamespace(namespace *types.Namespace) error {
	bytes, err := json.Marshal(namespace)
	if err != nil {
		return err
	}

	res, err := client.R().SetBody(bytes).Put("/rbac/namespaces/" + url.PathEscape(namespace.Name))
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// DeleteNamespace deletes an namespace on configured Sensu instance
func (client *RestClient) DeleteNamespace(namespace string) error {
	res, err := client.R().Delete("/rbac/namespaces/" + url.PathEscape(namespace))

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// ListNamespaces fetches all namespaces from configured Sensu instance
func (client *RestClient) ListNamespaces() ([]types.Namespace, error) {
	var namespaces []types.Namespace

	res, err := client.R().Get("/rbac/namespaces")
	if err != nil {
		return namespaces, err
	}

	if res.StatusCode() >= 400 {
		return namespaces, fmt.Errorf("%v", res.String())
	}

	err = json.Unmarshal(res.Body(), &namespaces)
	return namespaces, err
}

// FetchNamespace fetches an namespace by name
func (client *RestClient) FetchNamespace(namespaceName string) (*types.Namespace, error) {
	var namespace *types.Namespace

	res, err := client.R().Get("/rbac/namespaces/" + url.PathEscape(namespaceName))
	if err != nil {
		return namespace, err
	}

	if res.StatusCode() >= 400 {
		return namespace, fmt.Errorf("error getting namespace: %v", res.String())
	}

	err = json.Unmarshal(res.Body(), &namespace)
	return namespace, err
}
