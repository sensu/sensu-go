package testing

// PostAPIKey ...
func (c *MockClient) PostAPIKey(path string, obj interface{}) (string, error) {
	args := c.Called(path, obj)
	return args.Get(0).(string), args.Error(1)
}
