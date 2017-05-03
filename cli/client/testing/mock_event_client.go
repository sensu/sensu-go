package testing

import "github.com/sensu/sensu-go/types"

func (c *MockClient) ListEvents() ([]types.Event, error) {
	args := c.Called()
	return args.Get(0).([]types.Event), args.Get(1).(error)
}
