package helpers

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

// With a default of 'json'
func TestAddFormatFlag(t *testing.T) {
	flags := &pflag.FlagSet{}

	AddFormatFlag(flags)

	formatFlag := flags.Lookup("format")
	assert.NotNil(t, formatFlag)
}
