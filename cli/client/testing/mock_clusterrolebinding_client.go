package testing

import "github.com/sensu/sensu-go/types"

// CreateClusterRoleBinding ...
func (c *MockClient) CreateClusterRoleBinding(obj *types.ClusterRoleBinding) error {
	args := c.Called(obj)
	return args.Error(0)
}

// FetchClusterRoleBinding ...
func (c *MockClient) FetchClusterRoleBinding(name string) (*types.ClusterRoleBinding, error) {
	args := c.Called(name)
	return args.Get(0).(*types.ClusterRoleBinding), args.Error(1)
}

// DeleteClusterRoleBinding ...
func (c *MockClient) DeleteClusterRoleBinding(name string) error {
	args := c.Called(name)
	return args.Error(0)
}

// ListClusterRoleBindings ...
func (c *MockClient) ListClusterRoleBindings() ([]types.ClusterRoleBinding, error) {
	args := c.Called()
	return args.Get(0).([]types.ClusterRoleBinding), args.Error(1)
}
