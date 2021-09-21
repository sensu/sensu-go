package pipeline

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// DeleteCommand deletes a pipeline
func DeleteCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete [PIPELINE]",
		Short:        "delete pipelines",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Delete pipeline via API
			pipeline := args[0]
			namespace := cli.Config.Namespace()

			if skipConfirm, _ := cmd.Flags().GetBool("skip-confirm"); !skipConfirm {
				if confirmed := helpers.ConfirmDeleteResource(pipeline, "pipeline"); !confirmed {
					fmt.Fprintln(cmd.OutOrStdout(), "Canceled")
					return nil
				}
			}

			err := cli.Client.DeletePipeline(namespace, pipeline)
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
