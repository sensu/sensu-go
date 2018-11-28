package rolebinding

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// CreateCommand defines a new command to create a role binding
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create [NAME] --cluster-role=NAME|--role=NAME [--user=username] [--group=groupname]",
		Short:        "create a new role binding for a particular role or cluster role",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := helpers.VerifyName(args); err != nil {
				_ = cmd.Help()
				return err
			}

			var namespace string
			if namespace = helpers.GetChangedStringValueFlag("namespace", cmd.Flags()); namespace == "" {
				namespace = cli.Config.Namespace()
			}
			roleBinding := v2.NewRoleBinding(v2.NewObjectMeta(args[0], namespace))

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
				roleBinding.RoleRef = v2.RoleRef{
					Type: "ClusterRole",
					Name: clusterRole,
				}
			} else if role != "" {
				roleBinding.RoleRef = v2.RoleRef{
					Type: "Role",
					Name: role,
				}
			} else {
				return errors.New("a role or cluster role must be provided")
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
					v2.Subject{
						Type: "Group",
						Name: group,
					},
				)
			}
			for _, user := range users {
				roleBinding.Subjects = append(roleBinding.Subjects,
					v2.Subject{
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
		"the cluster role this role binding should reference",
	)
	_ = cmd.Flags().StringP("role", "r", "",
		"the role this role binding should reference",
	)
	_ = cmd.Flags().StringSliceP("group", "g", []string{},
		"groups to bind to the role",
	)
	_ = cmd.Flags().StringSliceP("user", "u", []string{},
		"users to bind to the role",
	)

	return cmd
}
