package filter

import (
	"github.com/sensu/sensu-go/cli"
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

		// TODO:(echlebek): add these back when the time window facility works
		// properly.
		//subcommands.RemoveWhenCommand(cli),
		//subcommands.SetWhenCommand(cli),
	)

	return cmd
}
