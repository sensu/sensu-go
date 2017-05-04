package testing

import "github.com/sensu/sensu-go/types"

func (c *MockClient) ListChecks() ([]types.Check, error) {
	args := c.Called()
	return args.Get(0).([]types.Check), args.Error(1)
}

func (c *MockClient) CreateCheck(check *types.Check) error {
	args := c.Called(check)
	return args.Error(0)
}

func (c *MockClient) DeleteCheck(check *types.Check) error {
	args := c.Called(check)
	return args.Error(0)
}
