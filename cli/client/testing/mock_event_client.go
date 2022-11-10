package testing

import (
	corev2 "github.com/sensu/core/v2"
)

// FetchEvent for use with mock lib
func (c *MockClient) FetchEvent(entity, check string) (*corev2.Event, error) {
	args := c.Called(entity, check)
	return args.Get(0).(*corev2.Event), args.Error(1)
}

// DeleteEvent for use with mock lib
func (c *MockClient) DeleteEvent(namespace, entity, check string) error {
	args := c.Called(namespace, entity, check)
	return args.Error(0)
}

// UpdateEvent for use with mock lib
func (c *MockClient) UpdateEvent(event *corev2.Event) error {
	args := c.Called(event)
	return args.Error(0)
}

// ResolveEvent for use with mock lib
func (c *MockClient) ResolveEvent(event *corev2.Event) error {
	args := c.Called(event)
	return args.Error(0)
}
