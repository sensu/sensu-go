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

func TestSafeSplitCSV(t *testing.T) {
	assert := assert.New(t)

	res := SafeSplitCSV("one")
	assert.Equal(res, []string{"one"})

	res = SafeSplitCSV("one,two")
	assert.Equal(res, []string{"one", "two"})

	res = SafeSplitCSV("one   ,       two")
	assert.Equal(res, []string{"one", "two"})

	res = SafeSplitCSV("one,     \t \u00a0 two")
	assert.Equal(res, []string{"one", "two"})

	res = SafeSplitCSV("    one ,     \t 🐛 two")
	assert.Equal(res, []string{"one", "🐛 two"})
}
