package helpers

import (
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/spf13/pflag"
)

// AddFormatFlag adds the format flag to the given command. When given client
// configuration the user's configured default format is used as the flag's
// default value.
func AddFormatFlag(flags *pflag.FlagSet) {
	flags.String("format", config.DefaultFormat, `format of data returned ("json"|"tabular")`)
}
