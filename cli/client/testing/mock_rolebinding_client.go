package testing

import "github.com/sensu/sensu-go/types"

// CreateRoleBinding ...
func (c *MockClient) CreateRoleBinding(obj *types.RoleBinding) error {
	args := c.Called(obj)
	return args.Error(0)
}

// FetchRoleBinding ...
func (c *MockClient) FetchRoleBinding(name string) (*types.RoleBinding, error) {
	args := c.Called(name)
	return args.Get(0).(*types.RoleBinding), args.Error(1)
}

// DeleteRoleBinding ...
func (c *MockClient) DeleteRoleBinding(name string) error {
	args := c.Called(name)
	return args.Error(0)
}

// ListRoleBindings ...
func (c *MockClient) ListRoleBindings() ([]types.RoleBinding, error) {
	args := c.Called()
	return args.Get(0).([]types.RoleBinding), args.Error(1)
}
