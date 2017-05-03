package testing

import (
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/mock"
)

type MockConfig struct {
	mock.Mock
}

func (c *MockConfig) Get(key string) interface{} {
	args := c.Called(key)
	return args.Get(0)
}
func (c *MockConfig) GetString(key string) string {
	args := c.Called(key)
	return args.Get(0).(string)
}

func (c *MockConfig) BindPFlag(key string, flag *pflag.Flag) {
	return
}
