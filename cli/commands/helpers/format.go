package helpers

import (
	"strings"

	"github.com/sensu/sensu-go/cli/client"
	"github.com/spf13/pflag"
)

func AddFormatFlag(flags *pflag.FlagSet, config client.Config) {
	defaultFormat := config.GetString("format")

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
