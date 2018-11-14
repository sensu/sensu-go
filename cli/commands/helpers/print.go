package helpers

import (
	"fmt"
	"io"

	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

type printTableFunc func(interface{}, io.Writer)

// Print displays
func Print(cmd *cobra.Command, format string, printTable printTableFunc, objects []types.Resource, v interface{}) error {
	if f := GetChangedStringValueFlag(flags.Format, cmd.Flags()); f != "" {
		format = f
	}
	switch format {
	case config.FormatJSON:
		return PrintJSON(v, cmd.OutOrStdout())
	case config.FormatWrappedJSON:
		if objects == nil {
			return PrintJSON(v, cmd.OutOrStdout())
		}
		return PrintWrappedJSONList(objects, cmd.OutOrStdout())
	case config.FormatYAML:
		if objects == nil {
			return PrintYAML(v, cmd.OutOrStdout())
		}
		return PrintYAML(objects, cmd.OutOrStdout())
	default:
		printTable(v, cmd.OutOrStdout())
	}

	return nil
}

// PrintFormatted prints the provided interface in the specified format.
// flag overrides the cli config format if set
func PrintFormatted(flag string, format string, v interface{}, w io.Writer, printToList func(interface{}, io.Writer) error) error {
	if flag != "" {
		format = flag
	}
	switch format {
	case config.FormatJSON:
		return PrintJSON(v, w)
	case config.FormatWrappedJSON:
		r, ok := v.(types.Resource)
		if !ok {
			return fmt.Errorf("%t is not a Resource", v)
		}
		return PrintWrappedJSON(r, w)
	case config.FormatYAML:
		return PrintYAML(v, w)
	default:
		return printToList(v, w)
	}
}
