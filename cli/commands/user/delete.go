package user

import (
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// DeleteCommand adds a command that allows admin's to disable users
func DeleteCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := cobra.Command{
		Use:          "disable USERNAME",
		Short:        "disable user given username",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 1 {
				cmd.Help()
				return nil
			}

			username := args[0]
			if skipConfirm, _ := cmd.Flags().GetBool("skip-confirm"); !skipConfirm {
				if confirmed := helpers.ConfirmDelete(username, cmd.OutOrStdout()); !confirmed {
					fmt.Fprintln(cmd.OutOrStdout(), "Canceled")
					return nil
				}
			}

			err := cli.Client.DisableUser(username)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Disabled")
			return nil
		},
	}

	cmd.Flags().Bool("skip-confirm", false, "skip interactive confirmation prompt")

	return &cmd
}
