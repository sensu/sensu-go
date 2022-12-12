package testing

import (
	corev2 "github.com/sensu/core/v2"
)

// CreateHook for use with mock lib
func (c *MockClient) CreateHook(hook *corev2.HookConfig) error {
	args := c.Called(hook)
	return args.Error(0)
}

// UpdateHook for use with mock lib
func (c *MockClient) UpdateHook(hook *corev2.HookConfig) error {
	args := c.Called(hook)
	return args.Error(0)
}

// DeleteHook for use with mock lib
func (c *MockClient) DeleteHook(namespace, name string) error {
	args := c.Called(namespace, name)
	return args.Error(0)
}

// FetchHook for use with mock lib
func (c *MockClient) FetchHook(name string) (*corev2.HookConfig, error) {
	args := c.Called(name)
	return args.Get(0).(*corev2.HookConfig), args.Error(1)
}
