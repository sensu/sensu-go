package handler

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// DeleteCommand adds a command that allows user to delete handlers
func DeleteCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete [NAME]",
		Short:        "delete handlers given name",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			name := args[0]
			namespace := cli.Config.Namespace()

			if skipConfirm, _ := cmd.Flags().GetBool("skip-confirm"); !skipConfirm {
				if confirmed := helpers.ConfirmDeleteResource(name, "handler"); !confirmed {
					fmt.Fprintln(cmd.OutOrStdout(), "Canceled")
					return nil
				}
			}

			err := cli.Client.DeleteHandler(namespace, name)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Deleted")
			return err
		},
	}

	cmd.Flags().Bool("skip-confirm", false, "skip interactive confirmation prompt")

	return cmd
}
