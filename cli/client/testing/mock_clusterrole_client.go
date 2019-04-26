package testing

import (
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

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
func (c *MockClient) ListClusterRoles(options client.ListOptions) ([]corev2.ClusterRole, error) {
	args := c.Called(options)
	return args.Get(0).([]corev2.ClusterRole), args.Error(1)
}
