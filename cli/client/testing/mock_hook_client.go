package testing

import "github.com/sensu/sensu-go/types"

// CreateHook for use with mock lib
func (c *MockClient) CreateHook(hook *types.HookConfig) error {
	args := c.Called(hook)
	return args.Error(0)
}

// UpdateHook for use with mock lib
func (c *MockClient) UpdateHook(hook *types.HookConfig) error {
	args := c.Called(hook)
	return args.Error(0)
}

// DeleteHook for use with mock lib
func (c *MockClient) DeleteHook(namespace, name string) error {
	args := c.Called(namespace, name)
	return args.Error(0)
}

// FetchHook for use with mock lib
func (c *MockClient) FetchHook(name string) (*types.HookConfig, error) {
	args := c.Called(name)
	return args.Get(0).(*types.HookConfig), args.Error(1)
}

// ListHooks for use with mock lib
func (c *MockClient) ListHooks(namespace string) ([]types.HookConfig, error) {
	args := c.Called(namespace)
	return args.Get(0).([]types.HookConfig), args.Error(1)
}
