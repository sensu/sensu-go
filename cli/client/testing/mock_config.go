package testing

import (
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

// BindPFlag for use with mock package
func (c *MockConfig) BindPFlag(key string, flag *pflag.Flag) {
	return
}
