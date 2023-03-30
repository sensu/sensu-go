package testing

import corev2 "github.com/sensu/core/v2"

// PostAPIKey ...
func (c *MockClient) PostAPIKey(path string, obj interface{}) (corev2.APIKeyResponse, error) {
	args := c.Called(path, obj)
	return args.Get(0).(corev2.APIKeyResponse), args.Error(1)
}
