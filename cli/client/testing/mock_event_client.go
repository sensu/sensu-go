package testing

import (
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// FetchEvent for use with mock lib
func (c *MockClient) FetchEvent(entity, check string) (*types.Event, error) {
	args := c.Called(entity, check)
	return args.Get(0).(*types.Event), args.Error(1)
}

// ListEvents for use with mock lib
func (c *MockClient) ListEvents(namespace string, options *client.ListOptions) ([]corev2.Event, error) {
	args := c.Called(namespace, options)
	return args.Get(0).([]corev2.Event), args.Error(1)
}

// DeleteEvent for use with mock lib
func (c *MockClient) DeleteEvent(namespace, entity, check string) error {
	args := c.Called(namespace, entity, check)
	return args.Error(0)
}

// UpdateEvent for use with mock lib
func (c *MockClient) UpdateEvent(event *types.Event) error {
	args := c.Called(event)
	return args.Error(0)
}

// ResolveEvent for use with mock lib
func (c *MockClient) ResolveEvent(event *types.Event) error {
	args := c.Called(event)
	return args.Error(0)
}
