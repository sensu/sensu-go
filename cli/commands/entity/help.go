package entity

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "entity",
		Short: "Manage entities",
	}

	// Add sub-commands
	cmd.AddCommand(
		ListCommand(cli),
		ShowCommand(cli),
		DeleteCommand(cli),
	)

	return cmd
}
