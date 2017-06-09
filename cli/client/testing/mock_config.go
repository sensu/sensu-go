package testing

import (
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/mock"
)

// MockConfig uses mock package to allow your tests to easily
// mock out the current Sensu client configuration
type MockConfig struct {
	mock.Mock
}

// Get for use with mock package
func (c *MockConfig) Get(key string) interface{} {
	args := c.Called(key)
	return args.Get(0)
}

// GetString for use with mock package
func (c *MockConfig) GetString(key string) string {
	args := c.Called(key)
	return args.Get(0).(string)
}

// GetTime for use with mock package
func (c *MockConfig) GetTime(key string) time.Time {
	args := c.Called(key)
	return args.Get(0).(time.Time)
}

// WriteURL for use with mock package
func (c *MockConfig) WriteURL(URL string) error {
	args := c.Called(URL)
	return args.Error(0)
}

// WriteCredentials for use with mock package
func (c *MockConfig) WriteCredentials(tokens *types.Tokens) error {
	args := c.Called(tokens)
	return args.Error(0)
}

// BindPFlag for use with mock package
func (c *MockConfig) BindPFlag(key string, flag *pflag.Flag) {
	return
}
