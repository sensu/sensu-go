package helpers

import (
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
	default:
		printTable(v, cmd.OutOrStdout())
	}

	return nil
}
