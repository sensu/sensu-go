package clusterrole

import (
	"errors"
	"fmt"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// CreateCommand defines new command to create a cluster role
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:		"create [NAME] --verbs=VERBS --resources=RESOURCES [--resource-name=RESOURCE_NAMES]",
		Short:		"create a new cluster role with a single rule",
		SilenceUsage:	true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := helpers.VerifyName(args); err != nil {
				_ = cmd.Help()
				return err
			}

			clusterRole := &v2.ClusterRole{
				ObjectMeta: v2.ObjectMeta{
					Name: args[0],
				},
			}

			// Retrieve the rule from the flags
			rule := v2.Rule{}

			verbs, err := cmd.Flags().GetStringSlice("verbs")
			if err != nil {
				return err
			}
			if len(verbs) == 0 {
				// Check the old "verb"
				verbs, err = cmd.Flags().GetStringSlice("verb")
				if err != nil {
					return err
				}
				// If it's still zero raise an error
				if len(verbs) == 0 {
					return errors.New("at least one verb must be provided")
				}
			}
			rule.Verbs = verbs

			resources, err := cmd.Flags().GetStringSlice("resources")
			if err != nil {
				return err
			}
			if len(resources) == 0 {
				// Check old resource
				resources, err = cmd.Flags().GetStringSlice("resource")
				if err != nil {
					return err
				}
				if len(resources) == 0 {
					return errors.New("at least one resource must be provided")
				}
			}
			rule.Resources = resources

			resourceNames, err := cmd.Flags().GetStringSlice("resource-name")
			if err != nil {
				return err
			}
			rule.ResourceNames = resourceNames

			// Assign the rule to our cluster role and validate it
			clusterRole.Rules = []v2.Rule{rule}
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

	// Non plural
	// To be removed in a later version?
	_ = cmd.Flags().StringSliceP("verb", "", []string{},
		"verbs that apply to the resources contained in the rule",
	)
	_ = cmd.Flags().StringSliceP("resource", "", []string{},
		"resources that the rule applies to",
	)

	// Plural
	_ = cmd.Flags().StringSliceP("verbs", "v", []string{},
		"verbs that apply to the resources contained in the rule",
	)
	_ = cmd.Flags().StringSliceP("resources", "r", []string{},
		"resources that the rule applies to",
	)
	_ = cmd.Flags().StringSliceP("resource-name", "n", []string{},
		"optional resource names that the rule applies to",
	)

	cmd.Flags().MarkDeprecated("resource", "please use resources instead.")
	cmd.Flags().MarkDeprecated("verb", "please use verbs instead.")
	cmd.Flags().MarkHidden("resource")
	cmd.Flags().MarkHidden("verb")

	return cmd
}
