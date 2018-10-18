package testing

import "github.com/sensu/sensu-go/types"

// AddGroupToUser for use with mock lib
func (c *MockClient) AddGroupToUser(username, group string) error {
	args := c.Called(username, group)
	return args.Error(0)
}

// AddRoleToUser for use with mock lib
func (c *MockClient) AddRoleToUser(username, role string) error {
	args := c.Called(username, role)
	return args.Error(0)
}

// CreateUser for use with mock lib
func (c *MockClient) CreateUser(user *types.User) error {
	args := c.Called(user)
	return args.Error(0)
}

// DisableUser for use with mock lib
func (c *MockClient) DisableUser(username string) error {
	args := c.Called(username)
	return args.Error(0)
}

// ListUsers for use with mock lib
func (c *MockClient) ListUsers() ([]types.User, error) {
	args := c.Called()
	return args.Get(0).([]types.User), args.Error(1)
}

// ReinstateUser for use with mock lib
func (c *MockClient) ReinstateUser(uname string) error {
	args := c.Called(uname)
	return args.Error(0)
}

// RemoveAllGroupsFromUser for use with mock lib
func (c *MockClient) RemoveAllGroupsFromUser(username string) error {
	args := c.Called(username)
	return args.Error(0)
}

// RemoveGroupFromUser for use with mock lib
func (c *MockClient) RemoveGroupFromUser(username, group string) error {
	args := c.Called(username, group)
	return args.Error(0)
}

// RemoveRoleFromUser for use with mock lib
func (c *MockClient) RemoveRoleFromUser(username, role string) error {
	args := c.Called(username, role)
	return args.Error(0)
}

// SetGroupsForUser for use with mock lib
func (c *MockClient) SetGroupsForUser(username string, groups []string) error {
	args := c.Called(username, groups)
	return args.Error(0)
}

// UpdatePassword for use with mock lib
func (c *MockClient) UpdatePassword(username, pwd string) error {
	args := c.Called(username, pwd)
	return args.Error(0)
}
