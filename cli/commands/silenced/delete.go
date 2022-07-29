package silenced

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
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
			namespace := cli.Config.Namespace()

			if skipConfirm, _ := cmd.Flags().GetBool("skip-confirm"); !skipConfirm {
				opts := []survey.AskOpt{
					survey.WithStdio(cli.InFile, cli.OutFile, cli.ErrFile),
				}
				if confirmed := helpers.ConfirmDeleteResourceWithOpts(name, "silence", opts...); !confirmed {
					fmt.Fprintln(cmd.OutOrStdout(), "Canceled")
					return nil
				}
			}

			err = cli.Client.DeleteSilenced(namespace, name)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Deleted")
			return err
		},
	}

	cmd.Flags().Bool("skip-confirm", false, "skip interactive confirmation prompt")
	cmd.Flags().StringP("subscription", "s", "", "silenced subscription")
	cmd.Flags().StringP("check", "c", "", "silenced check")

	return cmd
}
