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
		DeleteCommand(cli),
		ListCommand(cli),
		ShowCommand(cli),
		UpdateCommand(cli),
	)

	return cmd
}
