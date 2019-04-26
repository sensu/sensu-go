package helpers

import (
	"testing"

	"github.com/sensu/sensu-go/cli/commands/flags"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

// With a default of 'json'
func TestAddFormatFlag(t *testing.T) {
	flagSet := &pflag.FlagSet{}

	AddFormatFlag(flagSet)

	formatFlag := flagSet.Lookup("format")
	assert.NotNil(t, formatFlag)
}

func TestAddFieldSelectorFlag(t *testing.T) {
	flagSet := &pflag.FlagSet{}

	AddFieldSelectorFlag(flagSet)

	fieldSelectorFlag := flagSet.Lookup(flags.FieldSelector)
	assert.NotNil(t, fieldSelectorFlag)
}

func TestAddLabelSelectorFlag(t *testing.T) {
	flagSet := &pflag.FlagSet{}

	AddLabelSelectorFlag(flagSet)

	labelSelectorFlag := flagSet.Lookup(flags.LabelSelector)
	assert.NotNil(t, labelSelectorFlag)
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
