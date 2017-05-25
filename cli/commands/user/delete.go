package user

import (
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// DeleteCommand adds a command that allows user to delete users
func DeleteCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "delete [USERNAME]",
		Short:        "delete user given username",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 1 {
				cmd.Help()
				return nil
			}

			username := args[0]
			err := cli.Client.DeleteUser(username)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Deleted")
			return nil
		},
	}
}
