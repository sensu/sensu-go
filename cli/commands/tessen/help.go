package tessen

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tessen",
		Short: "Manage tessen configuration",
	}

	// Add sub-commands
	cmd.AddCommand(
		OptInCommand(cli),
		OptOutCommand(cli),
		InfoCommand(cli),
	)

	return cmd
}
