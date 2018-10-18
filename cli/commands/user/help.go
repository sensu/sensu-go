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
		AddGroupCommand(cli),
		AddRoleCommand(cli),
		CreateCommand(cli),
		DeleteCommand(cli),
		ListCommand(cli),
		ReinstateCommand(cli),
		RemoveGroupCommand(cli),
		RemoveAllGroupsCommand(cli),
		RemoveRoleCommand(cli),
		SetGroupsCommand(cli),
		SetPasswordCommand(cli),
	)

	return cmd
}
