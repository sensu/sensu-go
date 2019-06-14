package testing

// FetchClusterID ...
func (c *MockClient) FetchClusterID() (string, error) {
	args := c.Called()
	return args.Get(0).(string), args.Error(1)
}
