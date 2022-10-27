package testing

import (
	corev2 "github.com/sensu/core/v2"
)

// FetchPipeline for use with mock lib
func (c *MockClient) FetchPipeline(name string) (*corev2.Pipeline, error) {
	args := c.Called(name)
	return args.Get(0).(*corev2.Pipeline), args.Error(1)
}

// DeletePipeline for use with mock lib
func (c *MockClient) DeletePipeline(namespace, name string) error {
	args := c.Called(namespace, name)
	return args.Error(0)
}

// UpdatePipeline for use with mock lib
func (c *MockClient) UpdatePipeline(pipeline *corev2.Pipeline) error {
	args := c.Called(pipeline)
	return args.Error(0)
}

// CreatePipeline for use with mock lib
func (c *MockClient) CreatePipeline(pipeline *corev2.Pipeline) error {
	args := c.Called(pipeline)
	return args.Error(0)
}
