package user

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// SetGroupsCommand adds a command that allows admins to set the groups for a user.
func SetGroupsCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "set-groups USERNAME GROUP1[,GROUP2, ...[,GROUPN]]",
		Short:        "set the groups for a given user",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 2 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			username := args[0]
			groups := helpers.SafeSplitCSV(args[1])
			if err := cli.Client.SetGroupsForUser(username, groups); err != nil {
				return err
			}

			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Set")
			return err
		},
	}
}
