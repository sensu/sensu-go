package helpers

import (
	"strings"

	clientconfig "github.com/sensu/sensu-go/cli/client/config"
	"github.com/spf13/pflag"
)

// AddFormatFlag adds the format flag to the given command. When given client
// configuration the user's configured default format is used as the flag's
// default value.
func AddFormatFlag(flags *pflag.FlagSet, config clientconfig.Config) {
	defaultFormat := config.GetString("format")

	// Ensure that the configured default is trimmed and all lowercase
	// to match our command's expectations of a format.
	switch t := strings.ToLower(strings.TrimSpace(defaultFormat)); t {
	case "yaml":
		fallthrough
	case "json":
		defaultFormat = t
	default:
		defaultFormat = "tabular"
	}

	flags.StringP("format", "", defaultFormat, "format of data returned")
}
