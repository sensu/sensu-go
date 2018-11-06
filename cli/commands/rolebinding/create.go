package rolebinding

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// CreateCommand defines a new command to create a role binding
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create [NAME]",
		Short:        "create a new role binding",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("a name is required")
			}

			roleBinding := &types.RoleBinding{Name: args[0]}
			if namespace := helpers.GetChangedStringValueFlag("namespace", cmd.Flags()); namespace != "" {
				roleBinding.Namespace = namespace
			} else {
				roleBinding.Namespace = cli.Config.Namespace()
			}

			// Assign the rule to our role and valiate it
			if err := roleBinding.Validate(); err != nil {
				return err
			}

			if err := cli.Client.CreateRoleBinding(roleBinding); err != nil {
				return err
			}
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Created")
			return err
		},
	}

	return cmd
}
