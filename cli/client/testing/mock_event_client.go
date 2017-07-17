package testing

import "github.com/sensu/sensu-go/types"

// FetchEvent for use with mock lib
func (c *MockClient) FetchEvent(entity, check string) (*types.Event, error) {
	args := c.Called(entity, check)
	return args.Get(0).(*types.Event), args.Error(1)
}

// ListEvents for use with mock lib
func (c *MockClient) ListEvents() ([]types.Event, error) {
	args := c.Called()
	return args.Get(0).([]types.Event), args.Error(1)
}
