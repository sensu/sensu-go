package user

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// AddGroupCommand adds a command that allows admins to add a group to a user.
func AddGroupCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "add-group [USERNAME] [GROUP]",
		Short:        "add group to user given username and group",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 2 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			username := args[0]
			group := args[1]
			if err := cli.Client.AddGroupToUser(username, group); err != nil {
				return err
			}

			_, err := fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return err
		},
	}
}
