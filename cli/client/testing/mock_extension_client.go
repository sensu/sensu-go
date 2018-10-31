package testing

import "github.com/sensu/sensu-go/types"

// ListExtensions ...
func (c *MockClient) ListExtensions(namespace string) ([]types.Extension, error) {
	args := c.Called(namespace)
	return args.Get(0).([]types.Extension), args.Error(1)
}

// RegisterExtension ...
func (c *MockClient) RegisterExtension(e *types.Extension) error {
	args := c.Called(e)
	return args.Error(0)
}

// DeregisterExtension ...
func (c *MockClient) DeregisterExtension(name, namespace string) error {
	args := c.Called(name, namespace)
	return args.Error(0)
}
