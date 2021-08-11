package rolebinding

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// DeleteCommand defines new command to delete a role binding
func DeleteCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete [NAME]",
		Short:        "delete a role binding with the given name",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("a role binding name is required")
			}
			name := args[0]
			namespace := cli.Config.Namespace()

			if skipConfirm, _ := cmd.Flags().GetBool("skip-confirm"); !skipConfirm {
				if confirmed := helpers.ConfirmDeleteResource(name, "role-binding"); !confirmed {
					_, err := fmt.Fprintln(cmd.OutOrStdout(), "Canceled")
					return err
				}
			}

			err := cli.Client.DeleteRoleBinding(namespace, name)
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
