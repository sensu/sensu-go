package client

import (
	"encoding/json"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
)

var namespacesPath = CreateBasePath(coreAPIGroup, coreAPIVersion, "namespaces")

// CreateNamespace creates new namespace on configured Sensu instance
func (client *RestClient) CreateNamespace(namespace *types.Namespace) error {
	bytes, err := json.Marshal(namespace)
	if err != nil {
		return err
	}

	path := namespacesPath()
	res, err := client.R().SetBody(bytes).Post(path)

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// UpdateNamespace updates given namespace on a configured Sensu instance
func (client *RestClient) UpdateNamespace(namespace *types.Namespace) error {
	bytes, err := json.Marshal(namespace)
	if err != nil {
		return err
	}

	path := namespacesPath(namespace.Name)
	res, err := client.R().SetBody(bytes).Put(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// DeleteNamespace deletes an namespace on configured Sensu instance
func (client *RestClient) DeleteNamespace(namespace string) error {
	return client.Delete(namespacesPath(namespace))
}

// ListNamespaces fetches all namespaces from configured Sensu instance
func (client *RestClient) ListNamespaces(options *ListOptions) ([]corev2.Namespace, error) {
	var namespaces []corev2.Namespace

	if err := client.List(namespacesPath(), &namespaces, options); err != nil {
		return namespaces, err
	}

	return namespaces, nil
}

// FetchNamespace fetches an namespace by name
func (client *RestClient) FetchNamespace(namespaceName string) (*types.Namespace, error) {
	var namespace *types.Namespace

	path := namespacesPath(namespaceName)
	res, err := client.R().Get(path)
	if err != nil {
		return namespace, err
	}

	if res.StatusCode() >= 400 {
		return namespace, UnmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &namespace)
	return namespace, err
}
