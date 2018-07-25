package testing

// FetchLicense for use with mock lib
func (c *MockClient) FetchLicense() (interface{}, error) {
	args := c.Called()
	return args.Get(0), args.Error(1)
}

// UpdateLicense for use with mock lib
func (c *MockClient) UpdateLicense(license interface{}) error {
	args := c.Called(license)
	return args.Error(0)
}
