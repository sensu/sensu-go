package user

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// RemoveRoleCommand adds a command that allows admin's to remove a role from a user
func RemoveRoleCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "remove-role [USERNAME] [ROLE]",
		Short:        "remove role from user given username and role",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 2 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			username := args[0]
			role := args[1]
			if err := cli.Client.RemoveRoleFromUser(username, role); err != nil {
				return err
			}

			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Removed")
			return err
		},
	}
}
