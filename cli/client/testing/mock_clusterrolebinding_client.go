package testing

import (
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

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
func (c *MockClient) ListClusterRoleBindings(options client.ListOptions) ([]corev2.ClusterRoleBinding, string, error) {
	args := c.Called(options)
	return args.Get(0).([]corev2.ClusterRoleBinding), args.Get(1).(string), args.Error(2)
}
