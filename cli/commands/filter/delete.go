package filter

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// DeleteCommand defines the 'filter delete' subcommand
func DeleteCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete [NAME]",
		Short:        "delete filter given name",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			name := args[0]

			if skipConfirm, _ := cmd.Flags().GetBool("skip-confirm"); !skipConfirm {
				if confirmed := helpers.ConfirmDelete(name); !confirmed {
					fmt.Fprintln(cmd.OutOrStdout(), "Canceled")
					return nil
				}
			}

			filter := &types.EventFilter{Name: name}

			if org, _ := cmd.Flags().GetString("organization"); org != "" {
				filter.Organization = org
			}

			if env, _ := cmd.Flags().GetString("environment"); env != "" {
				filter.Environment = env
			}

			err := cli.Client.DeleteFilter(filter)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	_ = cmd.Flags().Bool("skip-confirm", false, "skip interactive confirmation prompt")

	return cmd
}
