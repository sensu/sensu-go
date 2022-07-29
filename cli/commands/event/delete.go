package event

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// DeleteCommand deletes an event
func DeleteCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete [ENTITY] [CHECK]",
		Short:        "delete events",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Delete event via API
			entity := args[0]
			check := args[1]
			namespace := cli.Config.Namespace()

			if skipConfirm, _ := cmd.Flags().GetBool("skip-confirm"); !skipConfirm {
				opts := []survey.AskOpt{
					survey.WithStdio(cli.InFile, cli.OutFile, cli.ErrFile),
				}
				name := fmt.Sprintf("%s/%s", entity, check)
				if confirmed := helpers.ConfirmDeleteResourceWithOpts(name, "event", opts...); !confirmed {
					fmt.Fprintln(cmd.OutOrStdout(), "Canceled")
					return nil
				}
			}

			err := cli.Client.DeleteEvent(namespace, entity, check)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Deleted")
			return err
		},
	}

	_ = cmd.Flags().Bool("skip-confirm", false, "skip interactive confirmation prompt")

	return cmd
}
