package client

import (
	"encoding/json"

	corev2 "github.com/sensu/core/v2"
)

// NamespacesPath is the api path for namespaces.
var NamespacesPath = CreateBasePath(coreAPIGroup, coreAPIVersion, "namespaces")

// CreateNamespace creates new namespace on configured Sensu instance
func (client *RestClient) CreateNamespace(namespace *corev2.Namespace) error {
	bytes, err := json.Marshal(namespace)
	if err != nil {
		return err
	}

	path := NamespacesPath()
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
func (client *RestClient) UpdateNamespace(namespace *corev2.Namespace) error {
	bytes, err := json.Marshal(namespace)
	if err != nil {
		return err
	}

	path := NamespacesPath(namespace.Name)
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
	return client.Delete(NamespacesPath(namespace))
}

// FetchNamespace fetches an namespace by name
func (client *RestClient) FetchNamespace(namespaceName string) (*corev2.Namespace, error) {
	var namespace *corev2.Namespace

	path := NamespacesPath(namespaceName)
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
