package helpers

import (
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/pflag"
)

func AddFormatFlag(flags *pflag.FlagSet, cli *cli.SensuCli) {
	defaultFormat := cli.Config.GetString("format")

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
