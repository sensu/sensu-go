package silenced

import (
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// DeleteCommand adds a command that allows user to delete silenceds
func DeleteCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete [ID]",
		Short:        "delete silenced, optionally by ID",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If pattern is wrong print out help
			if len(args) > 1 {
				cmd.Help()
				return nil
			}
			var id string
			if len(args) > 0 {
				id = args[0]
			}

			if len(id) == 0 {
				sub, err := cmd.Flags().GetString("subscription")
				if err != nil {
					return err
				}

				check, err := cmd.Flags().GetString("check")
				if err != nil {
					return err
				}

				id, err = types.SilencedID(sub, check)
				if err != nil {
					id, err = askID()
					if err != nil {
						return err
					}
				}
			}

			if skipConfirm, _ := cmd.Flags().GetBool("skip-confirm"); !skipConfirm {
				if confirmed := helpers.ConfirmDelete(id, cmd.OutOrStdout()); !confirmed {
					fmt.Fprintln(cmd.OutOrStdout(), "Canceled")
					return nil
				}
			}

			err := cli.Client.DeleteSilenced(id)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	cmd.Flags().Bool("skip-confirm", false, "skip interactive confirmation prompt")
	cmd.Flags().String("subscription", "", "silenced subscription")
	cmd.Flags().String("check", "", "silenced check")

	return cmd
}
