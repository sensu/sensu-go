package testing

import (
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// CreateNamespace for use with mock lib
func (c *MockClient) CreateNamespace(namespace *types.Namespace) error {
	args := c.Called(namespace)
	return args.Error(0)
}

// UpdateNamespace for use with mock lib
func (c *MockClient) UpdateNamespace(namespace *types.Namespace) error {
	args := c.Called(namespace)
	return args.Error(0)
}

// DeleteNamespace for use with mock lib
func (c *MockClient) DeleteNamespace(namespace string) error {
	args := c.Called(namespace)
	return args.Error(0)
}

// ListNamespaces for use with mock lib
func (c *MockClient) ListNamespaces(options client.ListOptions) ([]corev2.Namespace, string, error) {
	args := c.Called(options)
	return args.Get(0).([]corev2.Namespace), args.Get(1).(string), args.Error(2)
}

// FetchNamespace for use with mock lib
func (c *MockClient) FetchNamespace(namespace string) (*types.Namespace, error) {
	args := c.Called(namespace)
	return args.Get(0).(*types.Namespace), args.Error(1)
}
