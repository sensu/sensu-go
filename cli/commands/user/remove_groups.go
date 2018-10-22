package user

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// RemoveGroupsCommand adds a command that allows admins to remove all the groups for a user.
func RemoveAllGroupsCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "remove-groups USERNAME",
		Short:        "remove all the groups for a given user",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			username := args[0]
			if err := cli.Client.RemoveAllGroupsFromUser(username); err != nil {
				return err
			}

			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Removed")
			return err
		},
	}
}
