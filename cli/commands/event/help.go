package event

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// HelpCommand defines new event command
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event",
		Short: "Manage events",
	}

	// Add sub-commands
	cmd.AddCommand(ListCommand(cli))
	cmd.AddCommand(InfoCommand(cli))
	cmd.AddCommand(DeleteCommand(cli))
	cmd.AddCommand(ResolveCommand(cli))

	return cmd
}
