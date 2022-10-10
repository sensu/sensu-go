package testing

import (
	corev2 "github.com/sensu/core/v2"
)

// CreateNamespace for use with mock lib
func (c *MockClient) CreateNamespace(namespace *corev2.Namespace) error {
	args := c.Called(namespace)
	return args.Error(0)
}

// UpdateNamespace for use with mock lib
func (c *MockClient) UpdateNamespace(namespace *corev2.Namespace) error {
	args := c.Called(namespace)
	return args.Error(0)
}

// DeleteNamespace for use with mock lib
func (c *MockClient) DeleteNamespace(namespace string) error {
	args := c.Called(namespace)
	return args.Error(0)
}

// FetchNamespace for use with mock lib
func (c *MockClient) FetchNamespace(namespace string) (*corev2.Namespace, error) {
	args := c.Called(namespace)
	return args.Get(0).(*corev2.Namespace), args.Error(1)
}
