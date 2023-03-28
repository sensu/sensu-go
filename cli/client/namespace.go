package client

import (
	"encoding/json"

	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/types"
)

// NamespacesPath is the api path for namespaces.
var NamespacesPath = CreateBasePath("core", "v3", "namespaces")

// CreateNamespace creates new namespace on configured Sensu instance
func (client *RestClient) CreateNamespace(namespace *corev3.Namespace) error {
	bytes, err := json.Marshal(types.WrapResource(namespace))
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
func (client *RestClient) UpdateNamespace(namespace *corev3.Namespace) error {
	bytes, err := json.Marshal(types.WrapResource(namespace))
	if err != nil {
		return err
	}

	path := NamespacesPath(namespace.Metadata.Name)
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
func (client *RestClient) FetchNamespace(namespaceName string) (*corev3.Namespace, error) {
	path := NamespacesPath(namespaceName)
	res, err := client.R().Get(path)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, UnmarshalError(res)
	}

	var wrapper types.Wrapper
	err = json.Unmarshal(res.Body(), &wrapper)
	return wrapper.Value.(*corev3.Namespace), err
}
