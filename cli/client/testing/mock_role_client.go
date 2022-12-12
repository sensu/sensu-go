package testing

import (
	corev2 "github.com/sensu/core/v2"
)

// CreateRole ...
func (c *MockClient) CreateRole(check *corev2.Role) error {
	args := c.Called(check)
	return args.Error(0)
}

// FetchRole ...
func (c *MockClient) FetchRole(name string) (*corev2.Role, error) {
	args := c.Called(name)
	return args.Get(0).(*corev2.Role), args.Error(1)
}

// DeleteRole ...
func (c *MockClient) DeleteRole(namespace, name string) error {
	args := c.Called(namespace, name)
	return args.Error(0)
}
