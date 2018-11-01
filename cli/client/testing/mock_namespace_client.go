package testing

import "github.com/sensu/sensu-go/types"

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
func (c *MockClient) ListNamespaces() ([]types.Namespace, error) {
	args := c.Called()
	return args.Get(0).([]types.Namespace), args.Error(1)
}

// FetchNamespace for use with mock lib
func (c *MockClient) FetchNamespace(namespace string) (*types.Namespace, error) {
	args := c.Called(namespace)
	return args.Get(0).(*types.Namespace), args.Error(1)
}
