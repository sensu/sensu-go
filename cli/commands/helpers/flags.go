package helpers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/spf13/pflag"
)

var commaWhitespaceRegex *regexp.Regexp

// AddFormatFlag adds the format flag to the given command. When given client
// configuration the user's configured default format is used as the flag's
// default value.
func AddFormatFlag(flagSet *pflag.FlagSet) {
	flagSet.String(
		"format",
		config.DefaultFormat,
		fmt.Sprintf(
			`format of data returned ("%s"|"%s"|"%s")`,
			config.FormatJSON,
			config.FormatWrappedJSON,
			config.FormatTabular,
		),
	)
}

// AddAllNamespace adds the '--all-namespaces' flag to the given command
func AddAllNamespace(flagSet *pflag.FlagSet) {
	flagSet.Bool(flags.AllNamespaces, false, "Include records from all namespaces")
}

// AddInteractiveFlag adds the '--interactive' flag to the given command
func AddInteractiveFlag(flagSet *pflag.FlagSet) {
	flagSet.Bool(flags.Interactive, false, "Determines if CLI is in interactive mode")
}

// AddFieldSelectorFlag adds the '--field-selector' flag to the given command
func AddFieldSelectorFlag(flagSet *pflag.FlagSet) {
	flagSet.String(flags.FieldSelector, "", "Only select resources matching this field selector (enterprise only)")
}

// AddLabelSelectorFlag adds the '--label-selector' flag to the given command
func AddLabelSelectorFlag(flagSet *pflag.FlagSet) {
	flagSet.String(flags.LabelSelector, "", "Only select resources matching this label selector (enterprise only)")
}

// AddChunkSizeFlag adds the '--chunk-size' flag to the given command
func AddChunkSizeFlag(flagSet *pflag.FlagSet) {
	flagSet.Int(flags.ChunkSize, 0, "Return large lists in chunks of the given size rather than all at once")
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

// ListOptionsFromFlags construct an appropriate ListOptions given a FlagSet.
func ListOptionsFromFlags(flagSet *pflag.FlagSet) (client.ListOptions, error) {
	opts := client.ListOptions{}

	fieldSelector, err := flagSet.GetString(flags.FieldSelector)
	if err != nil {
		return opts, err
	}

	labelSelector, err := flagSet.GetString(flags.LabelSelector)
	if err != nil {
		return opts, err
	}

	chunkSize, err := flagSet.GetInt(flags.ChunkSize)
	if err != nil {
		chunkSize = 0
	}

	opts.FieldSelector = fieldSelector
	opts.LabelSelector = labelSelector
	opts.ChunkSize = chunkSize

	return opts, nil
}

func init() {
	// Matches same whitespace that the stdlib's unicode or strings packages would
	// https://golang.namespace/src/unicode/graphic.go?s=3997:4022#L116
	whiteSpc := "\\t\\n\\v\\f\\r\u0085\u00A0 "
	commaWhitespaceRegex = regexp.MustCompile(
		fmt.Sprintf("[%s]*,[%s]*", whiteSpc, whiteSpc),
	)
}
