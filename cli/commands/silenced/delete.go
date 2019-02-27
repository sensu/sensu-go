package silenced

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// DeleteCommand adds a command that allows user to delete silenceds
func DeleteCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete [NAME]",
		Short:        "delete silenced, optionally by name",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If pattern is wrong print out help
			if len(args) > 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}
			name, err := getName(cmd, args)
			if err != nil {
				return err
			}

			if skipConfirm, _ := cmd.Flags().GetBool("skip-confirm"); !skipConfirm {
				if confirmed := helpers.ConfirmDelete(name); !confirmed {
					fmt.Fprintln(cmd.OutOrStdout(), "Canceled")
					return nil
				}
			}

			if err := cli.Client.DeleteSilenced(name); err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Deleted")
			return nil
		},
	}

	cmd.Flags().Bool("skip-confirm", false, "skip interactive confirmation prompt")
	cmd.Flags().StringP("subscription", "s", "", "silenced subscription")
	cmd.Flags().StringP("check", "c", "", "silenced check")

	return cmd
}
