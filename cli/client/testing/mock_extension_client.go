package testing

import "github.com/sensu/sensu-go/types"

// ListExtensions ...
func (c *MockClient) ListExtensions(org string) ([]types.Extension, error) {
	args := c.Called(org)
	return args.Get(0).([]types.Extension), args.Error(1)
}

// RegisterExtension ...
func (c *MockClient) RegisterExtension(e *types.Extension) error {
	args := c.Called(e)
	return args.Error(0)
}

// DeregisterExtension ...
func (c *MockClient) DeregisterExtension(name, org string) error {
	args := c.Called(name, org)
	return args.Error(0)
}
