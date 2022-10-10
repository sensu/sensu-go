package testing

import (
	corev2 "github.com/sensu/core/v2"
)

// CreateClusterRoleBinding ...
func (c *MockClient) CreateClusterRoleBinding(obj *corev2.ClusterRoleBinding) error {
	args := c.Called(obj)
	return args.Error(0)
}

// FetchClusterRoleBinding ...
func (c *MockClient) FetchClusterRoleBinding(name string) (*corev2.ClusterRoleBinding, error) {
	args := c.Called(name)
	return args.Get(0).(*corev2.ClusterRoleBinding), args.Error(1)
}

// DeleteClusterRoleBinding ...
func (c *MockClient) DeleteClusterRoleBinding(name string) error {
	args := c.Called(name)
	return args.Error(0)
}
