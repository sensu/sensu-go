package testing

import (
	corev2 "github.com/sensu/core/v2"
)

// CreateClusterRole ...
func (c *MockClient) CreateClusterRole(obj *corev2.ClusterRole) error {
	args := c.Called(obj)
	return args.Error(0)
}

// FetchClusterRole ...
func (c *MockClient) FetchClusterRole(name string) (*corev2.ClusterRole, error) {
	args := c.Called(name)
	return args.Get(0).(*corev2.ClusterRole), args.Error(1)
}

// DeleteClusterRole ...
func (c *MockClient) DeleteClusterRole(name string) error {
	args := c.Called(name)
	return args.Error(0)
}
