package user

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// AddRoleCommand adds a command that allows admin's to add a role to a user
func AddRoleCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "add-role [USERNAME] [ROLE]",
		Short:        "add role to user given username and role",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 2 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			username := args[0]
			role := args[1]
			if err := cli.Client.AddRoleToUser(username, role); err != nil {
				return err
			}

			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Added")
			return err
		},
	}
}
