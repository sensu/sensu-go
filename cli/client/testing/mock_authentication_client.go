package testing

import "github.com/sensu/sensu-go/types"

// CreateAccessToken for use with mock lib
func (c *MockClient) CreateAccessToken(url, u, p string) (*types.Tokens, error) {
	args := c.Called(url, u, p)
	return args.Get(0).(*types.Tokens), args.Error(1)
}

// RefreshAccessToken for use with mock lib
func (c *MockClient) RefreshAccessToken(token string) (*types.Tokens, error) {
	args := c.Called(token)
	return args.Get(0).(*types.Tokens), args.Error(1)
}
