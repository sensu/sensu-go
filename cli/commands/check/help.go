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
		AddCheckHookCommand(cli),
		CreateCommand(cli),
		DeleteCommand(cli),
		ExecuteCommand(cli),
		ListCommand(cli),
		RemoveCheckHookCommand(cli),
		RemoveProxyRequestsCommand(cli),
		SetProxyRequestsCommand(cli),
		ShowCommand(cli),
		SubdueCommand(cli),
		UpdateCommand(cli),
	)

	return cmd
}
