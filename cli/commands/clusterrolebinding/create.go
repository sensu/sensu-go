package clusterrolebinding

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// CreateCommand defines a new command to create a cluster role binding
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create [NAME]",
		Short:        "create a new cluster role binding",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("a name is required")
			}

			clusterRoleBinding := &types.ClusterRoleBinding{Name: args[0]}

			// Assign the rule to our role and valiate it
			if err := clusterRoleBinding.Validate(); err != nil {
				return err
			}

			if err := cli.Client.CreateClusterRoleBinding(clusterRoleBinding); err != nil {
				return err
			}
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Created")
			return err
		},
	}

	return cmd
}
