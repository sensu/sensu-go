package clusterrole

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// CreateCommand defines new command to create a cluster role
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create [NAME]",
		Short:        "create a new cluster role and assign a rule to it",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("a name is required")
			}

			clusterRole := &types.ClusterRole{Name: args[0]}

			// Retrieve the rule from the flags
			rule := types.Rule{}

			verbs, err := cmd.Flags().GetStringSlice("verbs")
			if err != nil {
				return err
			}
			if len(verbs) == 0 {
				return errors.New("at least one verb must be provided")
			}
			rule.Verbs = verbs

			resources, err := cmd.Flags().GetStringSlice("resources")
			if err != nil {
				return err
			}
			if len(resources) == 0 {
				return errors.New("at least one resource must be provided")
			}
			rule.Resources = resources

			resourceNames, err := cmd.Flags().GetStringSlice("resource-names")
			if err != nil {
				return err
			}
			rule.ResourceNames = resourceNames

			// Assign the rule to our cluster role and validate it
			clusterRole.Rules = []types.Rule{rule}
			if err := clusterRole.Validate(); err != nil {
				return err
			}

			if err := cli.Client.CreateClusterRole(clusterRole); err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Created")
			return err
		},
	}

	_ = cmd.Flags().StringSliceP("verbs", "v", []string{},
		"list of verbs that apply to all of the listed resources for this rule",
	)
	_ = cmd.Flags().StringSliceP("resources", "r", []string{},
		"list of resources that this rule applies to",
	)
	_ = cmd.Flags().StringSliceP("resource-names", "n", []string{},
		"optional list of resource names that the rule applies to",
	)

	return cmd
}
