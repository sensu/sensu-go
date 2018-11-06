package testing

import "github.com/sensu/sensu-go/types"

// CreateClusterRole ...
func (c *MockClient) CreateClusterRole(obj *types.ClusterRole) error {
	args := c.Called(obj)
	return args.Error(0)
}

// FetchClusterRole ...
func (c *MockClient) FetchClusterRole(name string) (*types.ClusterRole, error) {
	args := c.Called(name)
	return args.Get(0).(*types.ClusterRole), args.Error(1)
}

// DeleteClusterRole ...
func (c *MockClient) DeleteClusterRole(name string) error {
	args := c.Called(name)
	return args.Error(0)
}

// ListClusterRoles ...
func (c *MockClient) ListClusterRoles() ([]types.ClusterRole, error) {
	args := c.Called()
	return args.Get(0).([]types.ClusterRole), args.Error(1)
}
