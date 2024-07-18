package testing

import corev2 "github.com/sensu/core/v2"

// DeletePipeline for use with mock lib
func (c *MockClient) DeleteFallbackPipeline(namespace, name string) error {
	args := c.Called(namespace, name)
	return args.Error(0)
}

// FetchFallbackPipeline for use with mock lib
func (c *MockClient) FetchFallbackPipeline(name string) (*corev2.FallbackPipeline, error) {
	args := c.Called(name)
	return args.Get(0).(*corev2.FallbackPipeline), args.Error(1)
}
