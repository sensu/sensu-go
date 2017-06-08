package testing

import creds "github.com/sensu/sensu-go/cli/client/credentials"

// CreateAccessToken for use with mock lib
func (c *MockClient) CreateAccessToken(url, u, p string) (*creds.AccessToken, error) {
	args := c.Called(url, u, p)
	return args.Get(0).(*creds.AccessToken), args.Error(1)
}

// RefreshAccessToken for use with mock lib
func (c *MockClient) RefreshAccessToken(token string) (*creds.AccessToken, error) {
	args := c.Called(token)
	return args.Get(0).(*creds.AccessToken), args.Error(1)
}
