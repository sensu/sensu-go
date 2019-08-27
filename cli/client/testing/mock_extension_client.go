package testing

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// RegisterExtension ...
func (c *MockClient) RegisterExtension(e *corev2.Extension) error {
	args := c.Called(e)
	return args.Error(0)
}

// DeregisterExtension ...
func (c *MockClient) DeregisterExtension(name, namespace string) error {
	args := c.Called(name, namespace)
	return args.Error(0)
}
