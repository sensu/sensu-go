package testing

import (
	"github.com/sensu/sensu-go/util/logging"
)

func (c *MockClient) SetLogLevel(logLevel logging.LogLevelRequest) error {
	args := c.Called(logLevel)
	return args.Error(0)
}

func (c *MockClient) SetLogLevelAllModules(logLevel string) error {
	args := c.Called(logLevel)
	return args.Error(0)
}

func (c *MockClient) GetLogLevel(logLevel string) error {
	args := c.Called(logLevel)
	return args.Error(0)
}
