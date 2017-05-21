package testing

import "github.com/sensu/sensu-go/types"

// CreateUser for use with mock lib
func (c *MockClient) CreateUser(check *types.User) error {
	args := c.Called(check)
	return args.Error(0)
}

// ListUsers for use with mock lib
func (c *MockClient) ListUsers() ([]types.User, error) {
	args := c.Called()
	return args.Get(0).([]types.User), args.Error(1)
}
