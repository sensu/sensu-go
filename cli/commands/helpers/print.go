package helpers

import (
	"io"

	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/spf13/cobra"
)

type printTableFunc func(interface{}, io.Writer)

// Print displays
func Print(cmd *cobra.Command, format string, printTable printTableFunc, objects interface{}) error {
	if f := GetChangedStringValueFlag(flags.Format, cmd.Flags()); f != "" {
		format = f
	}

	if format == "json" {
		if err := PrintJSON(objects, cmd.OutOrStdout()); err != nil {
			return err
		}
	} else {
		printTable(objects, cmd.OutOrStdout())
	}

	return nil
}
