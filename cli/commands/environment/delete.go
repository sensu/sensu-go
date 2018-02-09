package environment

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// DeleteCommand adds a command that allows user to delete environments
func DeleteCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := cobra.Command{
		Use:          "delete [ENVIRONMENT]",
		Short:        "delete specified environment",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 1 || args[0] == "" {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			org := cli.Config.Organization()
			env := args[0]
			if skipConfirm, _ := cmd.Flags().GetBool("skip-confirm"); !skipConfirm {
				if confirmed := helpers.ConfirmDelete(env); !confirmed {
					_, err := fmt.Fprintln(cmd.OutOrStdout(), "Canceled")
					return err
				}
			}

			err := cli.Client.DeleteEnvironment(org, env)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Deleted")
			return err
		},
	}

	_ = cmd.Flags().Bool("skip-confirm", false, "skip interactive confirmation prompt")

	return &cmd
}
