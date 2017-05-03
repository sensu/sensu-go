package check

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Manage checks",
	}

	// Add sub-commands
	cmd.AddCommand(
		ListCommand(cli),
		CreateCommand(cli),
		ImportCommand(cli),
		DeleteCommand(cli),
	)

	return cmd
}
