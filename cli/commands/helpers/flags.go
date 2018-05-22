package helpers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/spf13/pflag"
)

var commaWhitespaceRegex *regexp.Regexp

// AddFormatFlag adds the format flag to the given command. When given client
// configuration the user's configured default format is used as the flag's
// default value.
func AddFormatFlag(flagSet *pflag.FlagSet) {
	flagSet.String("format", config.DefaultFormat, `format of data returned ("json"|config.FormatTabular)`)
}

// AddAllOrganization adds the '--all-organizations' flag to the given command
func AddAllOrganization(flagSet *pflag.FlagSet) {
	flagSet.Bool(flags.AllOrgs, false, "Include records from all organizations")
}

// AddInteractiveFlag adds the '--interactive' flag to the given command
func AddInteractiveFlag(flagSet *pflag.FlagSet) {
	flagSet.Bool(flags.Interactive, false, "Determines if CLI is in interactive mode")
}

// FlagHasChanged determines if the user has set the value of a flag,
// or left it to default
func FlagHasChanged(name string, flagset *pflag.FlagSet) bool {
	flag := flagset.Lookup(name)
	if flag == nil {
		return false
	}
	return flag.Changed
}

// GetChangedStringValueFlag returns the value of a flag that has been explicitely
// changed by the user, and not left to default
func GetChangedStringValueFlag(name string, flagset *pflag.FlagSet) string {
	if !FlagHasChanged(name, flagset) {
		return ""
	}

	if value, err := flagset.GetString(name); err == nil {
		return value
	}

	return ""
}

// SafeSplitCSV splits given string and trims and extraneous whitespace
func SafeSplitCSV(i string) []string {
	trimmed := strings.TrimSpace(i)
	trimmed = commaWhitespaceRegex.ReplaceAllString(trimmed, ",")

	if len(trimmed) > 0 {
		return strings.Split(trimmed, ",")
	}

	return []string{}
}

func init() {
	// Matches same whitespace that the stdlib's unicode or strings packages would
	// https://golang.org/src/unicode/graphic.go?s=3997:4022#L116
	whiteSpc := "\\t\\n\\v\\f\\r\u0085\u00A0 "
	commaWhitespaceRegex = regexp.MustCompile(
		fmt.Sprintf("[%s]*,[%s]*", whiteSpc, whiteSpc),
	)
}
