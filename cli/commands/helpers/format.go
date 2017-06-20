package helpers

import (
	"errors"
	"strings"

	clientconfig "github.com/sensu/sensu-go/cli/client/config"
	"github.com/spf13/pflag"
)

// AddFormatFlag adds the format flag to the given command. When given client
// configuration the user's configured default format is used as the flag's
// default value.
func AddFormatFlag(flags *pflag.FlagSet, config clientconfig.Config) {
	defaultFormat := config.Format()

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

// JoinErrors joins multiple errors messages. Useful when
// you want the CLI to display more than one error message.
//
// eg.
//
//   JoinErrors("Validation: ", []error{errors.New("a"), errors.New("b")})
//   "Validation: a, and b."
func JoinErrors(prelude string, errs []error) error {
	out := prelude + " "
	lastElem := len(errs) - 1

	for i, err := range errs {
		var seperator string
		switch i {
		case 1:
			seperator = ""
		case lastElem:
			seperator = ", and "
		default:
			seperator = ", "
		}

		out += seperator + err.Error() + "."
	}

	return errors.New(out)
}
