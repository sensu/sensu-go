package asset

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset",
		Short: "Manage assets",
		RunE:  helpers.DefaultSubCommandRunE,
	}

	// Add sub-commands
	cmd.AddCommand(
		CreateCommand(cli),
		ListCommand(cli),
		InfoCommand(cli),
		UpdateCommand(cli),
		DeleteCommand(cli),
		AddCommand(cli),
		OutdatedCommand(cli),
	)
	return cmd
}
