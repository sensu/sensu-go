package user

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users",
	}

	// Add sub-commands
	cmd.AddCommand(
		AddRoleCommand(cli),
		CreateCommand(cli),
		DeleteCommand(cli),
		ListCommand(cli),
		ReinstateCommand(cli),
		RemoveRoleCommand(cli),
		SetPasswordCommand(cli),
		TestCredsCommand(cli),
	)

	return cmd
}
