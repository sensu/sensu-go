package testing

import "github.com/sensu/sensu-go/types"

// CreateRole ...
func (c *MockClient) CreateRole(check *types.Role) error {
	args := c.Called(check)
	return args.Error(0)
}

// FetchRole ...
func (c *MockClient) FetchRole(name string) (*types.Role, error) {
	args := c.Called(name)
	return args.Get(0).(*types.Role), args.Error(1)
}

// DeleteRole ...
func (c *MockClient) DeleteRole(name string) error {
	args := c.Called(name)
	return args.Error(0)
}

// ListRoles ...
func (c *MockClient) ListRoles(namespace string) ([]types.Role, error) {
	args := c.Called(namespace)
	return args.Get(0).([]types.Role), args.Error(1)
}
