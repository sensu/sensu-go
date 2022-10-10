package testing

import (
	corev2 "github.com/sensu/core/v2"
)

// CreateCheck for use with mock lib
func (c *MockClient) CreateCheck(check *corev2.CheckConfig) error {
	args := c.Called(check)
	return args.Error(0)
}

// UpdateCheck for use with mock lib
func (c *MockClient) UpdateCheck(check *corev2.CheckConfig) error {
	args := c.Called(check)
	return args.Error(0)
}

// DeleteCheck for use with mock lib
func (c *MockClient) DeleteCheck(namespace, name string) error {
	args := c.Called(namespace, name)
	return args.Error(0)
}

// ExecuteCheck for use with mock lib
func (c *MockClient) ExecuteCheck(req *corev2.AdhocRequest) error {
	args := c.Called(req)
	return args.Error(0)
}

// FetchCheck for use with mock lib
func (c *MockClient) FetchCheck(name string) (*corev2.CheckConfig, error) {
	args := c.Called(name)
	return args.Get(0).(*corev2.CheckConfig), args.Error(1)
}

// AddCheckHook for use with mock lib
func (c *MockClient) AddCheckHook(check *corev2.CheckConfig, checkHook *corev2.HookList) error {
	args := c.Called(check, checkHook)
	return args.Error(0)
}

// RemoveCheckHook for use with mock lib
func (c *MockClient) RemoveCheckHook(check *corev2.CheckConfig, hookType string, hookName string) error {
	args := c.Called(check, hookType, hookName)
	return args.Error(0)
}
