package testing

import "github.com/sensu/sensu-go/types"

// CreateAccessToken for use with mock lib
func (c *MockClient) CreateAccessToken(url, u, p string) (*types.Tokens, error) {
	args := c.Called(url, u, p)
	return args.Get(0).(*types.Tokens), args.Error(1)
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
func (c *MockClient) RefreshAccessToken(token string) (*types.Tokens, error) {
	args := c.Called(token)
	return args.Get(0).(*types.Tokens), args.Error(1)
}
