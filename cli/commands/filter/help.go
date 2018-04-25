package filter

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/filter/subcommands"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "filter",
		Short: "Manage filters",
	}

	// Add sub-commands
	cmd.AddCommand(
		CreateCommand(cli),
		DeleteCommand(cli),
		InfoCommand(cli),
		ListCommand(cli),
		UpdateCommand(cli),

		subcommands.RemoveWhenCommand(cli),
		subcommands.SetWhenCommand(cli),
	)

	return cmd
}
