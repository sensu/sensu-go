package user

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// RemoveGroupCommand adds a command that allows admins to remove a group from a user.
func RemoveGroupCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "remove-group [USERNAME] [group]",
		Short:        "remove group from user given username and group",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 2 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			username := args[0]
			group := args[1]
			if err := cli.Client.RemoveGroupFromUser(username, group); err != nil {
				return err
			}

			_, err := fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return err
		},
	}
}
