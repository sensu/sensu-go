package clusterrolebinding

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// CreateCommand defines a new command to create a cluster role binding
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create [NAME] --cluster-role=NAME [--user=username] [--group=groupname]",
		Short:        "create a new cluster role binding for a particular cluster role",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := helpers.VerifyName(args); err != nil {
				_ = cmd.Help()
				return err
			}

			clusterRoleBinding := &types.ClusterRoleBinding{
				ObjectMeta: v2.ObjectMeta{
					Name: args[0],
				},
			}

			clusterRole, err := cmd.Flags().GetString("cluster-role")
			if err != nil {
				return err
			}
			if clusterRole == "" {
				return errors.New("a ClusterRole must be provided")
			}
			clusterRoleBinding.RoleRef = types.RoleRef{
				Type: "ClusterRole",
				Name: clusterRole,
			}

			groups, err := cmd.Flags().GetStringSlice("group")
			if err != nil {
				return err
			}
			users, err := cmd.Flags().GetStringSlice("user")
			if err != nil {
				return err
			}
			if len(groups) == 0 && len(users) == 0 {
				return errors.New("at least one group or user must be provided")
			}

			// Create our subjects list
			for _, group := range groups {
				clusterRoleBinding.Subjects = append(clusterRoleBinding.Subjects,
					types.Subject{
						Type: "Group",
						Name: group,
					},
				)
			}
			for _, user := range users {
				clusterRoleBinding.Subjects = append(clusterRoleBinding.Subjects,
					types.Subject{
						Type: "User",
						Name: user,
					},
				)
			}

			// Assign the rule to our role and valiate it
			if err := clusterRoleBinding.Validate(); err != nil {
				return err
			}

			if err := cli.Client.CreateClusterRoleBinding(clusterRoleBinding); err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Created")
			return err
		},
	}

	_ = cmd.Flags().StringP("cluster-role", "c", "",
		"the cluster role this cluster role binding should reference",
	)
	_ = cmd.Flags().StringSliceP("group", "g", []string{},
		"groups to bind to the cluster role",
	)
	_ = cmd.Flags().StringSliceP("user", "u", []string{},
		"users to bind to the cluster role",
	)

	return cmd
}
