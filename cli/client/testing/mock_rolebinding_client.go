package testing

import (
	corev2 "github.com/sensu/core/v2"
)

// CreateRoleBinding ...
func (c *MockClient) CreateRoleBinding(obj *corev2.RoleBinding) error {
	args := c.Called(obj)
	return args.Error(0)
}

// FetchRoleBinding ...
func (c *MockClient) FetchRoleBinding(name string) (*corev2.RoleBinding, error) {
	args := c.Called(name)
	return args.Get(0).(*corev2.RoleBinding), args.Error(1)
}

// DeleteRoleBinding ...
func (c *MockClient) DeleteRoleBinding(namespace, name string) error {
	args := c.Called(namespace, name)
	return args.Error(0)
}
