package testing

import "github.com/sensu/sensu-go/types"

func (c *MockClient) ListChecks() ([]types.Check, error) {
	args := c.Called()
	return args.Get(0).([]types.Check), args.Get(1).(error)
}

func (c *MockClient) CreateCheck(check *types.Check) error {
	args := c.Called(check)
	return args.Get(0).(error)
}

func (c *MockClient) DeleteCheck(check *types.Check) error {
	args := c.Called(check)
	return args.Get(0).(error)
}
