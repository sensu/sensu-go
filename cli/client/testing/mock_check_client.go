package testing

import "github.com/sensu/sensu-go/types"

// CreateCheck for use with mock lib
func (c *MockClient) CreateCheck(check *types.CheckConfig) error {
	args := c.Called(check)
	return args.Error(0)
}

// UpdateCheck for use with mock lib
func (c *MockClient) UpdateCheck(check *types.CheckConfig) error {
	args := c.Called(check)
	return args.Error(0)
}

// DeleteCheck for use with mock lib
func (c *MockClient) DeleteCheck(check *types.CheckConfig) error {
	args := c.Called(check)
	return args.Error(0)
}

// ExecuteCheck for use with mock lib
func (c *MockClient) ExecuteCheck(req *types.AdhocRequest) error {
	args := c.Called(req)
	return args.Error(0)
}

// FetchCheck for use with mock lib
func (c *MockClient) FetchCheck(name string) (*types.CheckConfig, error) {
	args := c.Called(name)
	return args.Get(0).(*types.CheckConfig), args.Error(1)
}

// ListChecks for use with mock lib
func (c *MockClient) ListChecks(org string) ([]types.CheckConfig, error) {
	args := c.Called(org)
	return args.Get(0).([]types.CheckConfig), args.Error(1)
}

// AddCheckHook for use with mock lib
func (c *MockClient) AddCheckHook(check *types.CheckConfig, checkHook *types.HookList) error {
	args := c.Called(check, checkHook)
	return args.Error(0)
}

// RemoveCheckHook for use with mock lib
func (c *MockClient) RemoveCheckHook(check *types.CheckConfig, hookType string, hookName string) error {
	args := c.Called(check, hookType, hookName)
	return args.Error(0)
}
