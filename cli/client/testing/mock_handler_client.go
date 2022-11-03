package testing

import (
	corev2 "github.com/sensu/core/v2"
)

// CreateHandler for use with mock package
func (c *MockClient) CreateHandler(h *corev2.Handler) error {
	args := c.Called(h)
	return args.Error(0)
}

// DeleteHandler for use with mock package
func (c *MockClient) DeleteHandler(namespace, name string) error {
	args := c.Called(namespace, name)
	return args.Error(0)
}

// FetchHandler for use with mock lib
func (c *MockClient) FetchHandler(name string) (*corev2.Handler, error) {
	args := c.Called(name)
	return args.Get(0).(*corev2.Handler), args.Error(1)
}

// UpdateHandler for use with mock lib
func (c *MockClient) UpdateHandler(h *corev2.Handler) error {
	args := c.Called(h)
	return args.Error(0)
}
