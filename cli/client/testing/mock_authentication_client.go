package testing

import (
	corev2 "github.com/sensu/core/v2"
)

// CreateAccessToken for use with mock lib
func (c *MockClient) CreateAccessToken(url, u, p string) (*corev2.Tokens, error) {
	args := c.Called(url, u, p)
	return args.Get(0).(*corev2.Tokens), args.Error(1)
}

// TestCreds for use with mock lib
func (c *MockClient) TestCreds(u, p string) error {
	args := c.Called(u, p)
	return args.Error(0)
}

// Logout for use with mock lib
func (c *MockClient) Logout(token string) error {
	args := c.Called(token)
	return args.Error(0)
}

// RefreshAccessToken for use with mock lib
func (c *MockClient) RefreshAccessToken(tokens *corev2.Tokens) (*corev2.Tokens, error) {
	args := c.Called(tokens)
	return args.Get(0).(*corev2.Tokens), args.Error(1)
}
