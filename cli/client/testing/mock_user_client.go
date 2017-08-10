package testing

import "github.com/sensu/sensu-go/types"

// CreateUser for use with mock lib
func (c *MockClient) CreateUser(user *types.User) error {
	args := c.Called(user)
	return args.Error(0)
}

// UpdatePassword for use with mock lib
func (c *MockClient) UpdatePassword(username, pwd string) error {
	args := c.Called(username, pwd)
	return args.Error(0)
}

// DeleteUser for use with mock lib
func (c *MockClient) DeleteUser(username string) error {
	args := c.Called(username)
	return args.Error(0)
}

// ListUsers for use with mock lib
func (c *MockClient) ListUsers() ([]types.User, error) {
	args := c.Called()
	return args.Get(0).([]types.User), args.Error(1)
}
