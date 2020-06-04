package testing

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// AddGroupToUser for use with mock lib
func (c *MockClient) AddGroupToUser(username, group string) error {
	args := c.Called(username, group)
	return args.Error(0)
}

// CreateUser for use with mock lib
func (c *MockClient) CreateUser(user *corev2.User) error {
	args := c.Called(user)
	return args.Error(0)
}

// DisableUser for use with mock lib
func (c *MockClient) DisableUser(username string) error {
	args := c.Called(username)
	return args.Error(0)
}

// FetchUser for use with mock lib
func (c *MockClient) FetchUser(username string) (*corev2.User, error) {
	args := c.Called(username)
	return args.Get(0).(*corev2.User), args.Error(1)
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

// SetGroupsForUser for use with mock lib
func (c *MockClient) SetGroupsForUser(username string, groups []string) error {
	args := c.Called(username, groups)
	return args.Error(0)
}

// ResetPassword for use with mock lib
func (c *MockClient) ResetPassword(username, passwordHash string) error {
	args := c.Called(username, passwordHash)
	return args.Error(0)
}

// UpdatePassword for use with mock lib
func (c *MockClient) UpdatePassword(username, newPasswordHash, currentPassword string) error {
	args := c.Called(username, newPasswordHash, currentPassword)
	return args.Error(0)
}
