package testing

import "github.com/sensu/sensu-go/types"

// CreateRole for use with mock lib
func (c *MockClient) CreateRole(check *types.Role) error {
	args := c.Called(check)
	return args.Error(0)
}

// DeleteRole for use with mock lib
func (c *MockClient) DeleteRole(name string) error {
	args := c.Called(name)
	return args.Error(0)
}

// ListRoles for use with mock lib
func (c *MockClient) ListRoles() ([]types.Role, error) {
	args := c.Called()
	return args.Get(0).([]types.Role), args.Error(1)
}

// AddRule for use with mock lib
func (c *MockClient) AddRule(role string, rule *types.Rule) error {
	args := c.Called(role, rule)
	return args.Error(0)
}

// RemoveRule for use with mock lib
func (c *MockClient) RemoveRule(role string, ruleType string) error {
	args := c.Called(role, ruleType)
	return args.Error(0)
}
