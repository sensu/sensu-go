package fallbackPipeline

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// DeleteCommand deletes a pipeline
func DeleteCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete [FALLBACKPIPELINE]",
		Short:        "delete fallbackPipeline",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Delete pipeline via API
			fallbackPipeline := args[0]
			namespace := cli.Config.Namespace()

			if skipConfirm, _ := cmd.Flags().GetBool("skip-confirm"); !skipConfirm {
				opts := []survey.AskOpt{
					survey.WithStdio(cli.InFile, cli.OutFile, cli.ErrFile),
				}
				if confirmed := helpers.ConfirmDeleteResourceWithOpts(fallbackPipeline, "fallbackPipeline", opts...); !confirmed {
					fmt.Fprintln(cmd.OutOrStdout(), "Canceled")
					return nil
				}
			}

			//RAJ look here , should be DeleteFallbackPipeline
			err := cli.Client.DeletePipeline(namespace, fallbackPipeline)

			//cli.Client.Delete(cli.Client.FallbackPipelinesPath(namespace, name)

			//err := cli.Client.

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
