package role

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// CreateCommand defines new command to create roles
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create [NAME]",
		Short:        "create new roles",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			role := &types.Role{Name: args[0]}
			if err := role.Validate(); err != nil {
				return err
			}

			opts := &roleOpts{}

			opts.Namespace = cli.Config.Namespace()

			opts.withFlags(cmd.Flags())
			opts.Name = args[0]

			if err := cli.Client.CreateRole(role); err != nil {
				return err
			}
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Created")
			return err
		},
	}

	return cmd
}
