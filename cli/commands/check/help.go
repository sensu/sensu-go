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
		CreateCommand(cli),
		DeleteCommand(cli),
		ListCommand(cli),
		ShowCommand(cli),
		UpdateCommand(cli),
		AddCheckHookCommand(cli),
		RemoveCheckHookCommand(cli),
	)

	return cmd
}
