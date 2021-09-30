package command

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/cmdmanager"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// DeleteCommand adds command that allows a user to delete a command plugin.
func DeleteCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete [ALIAS]",
		Short:        "delete sensuctl commands",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			name := args[0]
			if skipConfirm, _ := cmd.Flags().GetBool("skip-confirm"); !skipConfirm {
				if confirmed := helpers.ConfirmDeleteResource(name, "command"); !confirmed {
					fmt.Fprintln(cmd.OutOrStdout(), "Canceled")
					return nil
				}
			}

			manager, err := cmdmanager.NewCommandManager(cli)
			if err != nil {
				return err
			}

			if err := manager.DeleteCommandPlugin(name); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Deleted")
			return err
		},
	}

	cmd.Flags().Bool("skip-confirm", false, "skip interactive confirmation prompt")

	return cmd
}
