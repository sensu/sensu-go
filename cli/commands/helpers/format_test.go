package helpers

import (
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

// With a default of 'json'
func TestAddFormatFlag(t *testing.T) {
	assert := assert.New(t)
	flags := &pflag.FlagSet{}

	config := &client.MockConfig{}
	config.On("GetString", "format").Return("json")

	AddFormatFlag(flags, config)

	formatFlag := flags.Lookup("format")
	assert.NotNil(formatFlag)
	assert.Equal("json", formatFlag.DefValue)
}

// With a default of 'JSon '
func TestAddFormatFlagWithBadChars(t *testing.T) {
	assert := assert.New(t)
	flags := &pflag.FlagSet{}

	config := &client.MockConfig{}
	config.On("GetString", "format").Return("JSon")

	AddFormatFlag(flags, config)

	formatFlag := flags.Lookup("format")
	assert.NotNil(formatFlag)
	assert.Equal("json", formatFlag.DefValue, "value is trimmed")
}

// With an no default ('')
func TestAddFormatFlagNoDefault(t *testing.T) {
	assert := assert.New(t)
	flags := &pflag.FlagSet{}

	config := &client.MockConfig{}
	config.On("GetString", "format").Return("")

	AddFormatFlag(flags, config)

	formatFlag := flags.Lookup("format")
	assert.NotNil(formatFlag)
	assert.Equal("tabular", formatFlag.DefValue, "falls back to tabular")
}

// With an unknown default ('blob')
func TestAddFormatFlagUnknownDefault(t *testing.T) {
	assert := assert.New(t)
	flags := &pflag.FlagSet{}

	config := &client.MockConfig{}
	config.On("GetString", "format").Return("blob")

	AddFormatFlag(flags, config)

	formatFlag := flags.Lookup("format")
	assert.NotNil(formatFlag)
	assert.Equal("tabular", formatFlag.DefValue, "falls back to tabular")
}
