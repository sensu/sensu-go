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
		Use:          "create [NAME] --cluster-role=NAME|--role=NAME [--user=username] [--group=groupname]",
		Short:        "create a new RoleBinding for a particular Role or ClusterRole",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := helpers.VerifyName(args); err != nil {
				_ = cmd.Help()
				return err
			}

			roleBinding := &types.RoleBinding{Name: args[0]}
			if namespace := helpers.GetChangedStringValueFlag("namespace", cmd.Flags()); namespace != "" {
				roleBinding.Namespace = namespace
			} else {
				roleBinding.Namespace = cli.Config.Namespace()
			}

			// Determine if a Role or ClusterRole was provided and assign it to our
			// RoleBinding
			role, err := cmd.Flags().GetString("role")
			if err != nil {
				return err
			}
			clusterRole, err := cmd.Flags().GetString("cluster-role")
			if err != nil {
				return err
			}

			if clusterRole != "" {
				roleBinding.RoleRef = types.RoleRef{
					Type: "ClusterRole",
					Name: clusterRole,
				}
			} else if role != "" {
				roleBinding.RoleRef = types.RoleRef{
					Type: "Role",
					Name: role,
				}
			} else {
				return errors.New("a Role or ClusterRole must be provided")
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
				roleBinding.Subjects = append(roleBinding.Subjects,
					types.Subject{
						Type: "Group",
						Name: group,
					},
				)
			}
			for _, user := range users {
				roleBinding.Subjects = append(roleBinding.Subjects,
					types.Subject{
						Type: "User",
						Name: user,
					},
				)
			}

			// Assign the rule to our role and valiate it
			if err := roleBinding.Validate(); err != nil {
				return err
			}

			if err := cli.Client.CreateRoleBinding(roleBinding); err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Created")
			return err
		},
	}

	_ = cmd.Flags().StringP("cluster-role", "c", "",
		"the ClusterRole this RoleBinding should reference",
	)
	_ = cmd.Flags().StringP("role", "r", "",
		"the Role this RoleBinding should reference",
	)
	_ = cmd.Flags().StringSliceP("group", "g", []string{},
		"groups to bind to the Role",
	)
	_ = cmd.Flags().StringSliceP("user", "u", []string{},
		"users to bind to the Role",
	)

	return cmd
}
