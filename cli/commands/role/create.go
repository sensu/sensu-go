package role

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// CreateCommand defines new command to create roles
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create [NAME]",
		Short:        "create new role",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			role := &types.Role{Name: args[0]}
			if namespace := helpers.GetChangedStringValueFlag("namespace", cmd.Flags()); namespace != "" {
				role.Namespace = namespace
			} else {
				role.Namespace = cli.Config.Namespace()
			}

			role.Rules = []types.Rule{
				types.FixtureRule(),
			}

			if err := role.Validate(); err != nil {
				return err
			}

			if err := cli.Client.CreateRole(role); err != nil {
				return err
			}
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Created")
			return err
		},
	}

	return cmd
}
