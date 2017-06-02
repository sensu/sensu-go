package testing

// CreateAccessToken for use with mock lib
func (c *MockClient) CreateAccessToken(string u, string p) (*client.AccessToken, error) {
	args := c.Called(u, p)
	return &args.Get(0).(AccessToken), args.Error(1)
}

// RefreshAccessToken for use with mock lib
func (c *MockClient) RefreshAccessToken(string token) (*client.AccessToken, error) {
	args := c.Called(token)
	return &args.Get(0).(AccessToken), args.Error(1)
}
