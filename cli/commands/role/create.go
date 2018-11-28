package role

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// CreateCommand defines new command to create roles
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create [NAME] --verb=VERBS --resource=RESOURCES [--resource-name=RESOURCE_NAMES]",
		Short:        "create a new role with a single rule",
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
			role := v2.NewRole(v2.NewObjectMeta(args[0], namespace))

			// Retrieve the rule from the flags
			rule := v2.Rule{}

			verbs, err := cmd.Flags().GetStringSlice("verb")
			if err != nil {
				return err
			}
			if len(verbs) == 0 {
				return errors.New("at least one verb must be provided")
			}
			rule.Verbs = verbs

			resources, err := cmd.Flags().GetStringSlice("resource")
			if err != nil {
				return err
			}
			if len(resources) == 0 {
				return errors.New("at least one resource must be provided")
			}
			rule.Resources = resources

			resourceNames, err := cmd.Flags().GetStringSlice("resource-name")
			if err != nil {
				return err
			}
			rule.ResourceNames = resourceNames

			// Assign the rule to our role and validate it
			role.Rules = []v2.Rule{rule}
			if err := role.Validate(); err != nil {
				return err
			}

			if err := cli.Client.CreateRole(role); err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Created")
			return err
		},
	}

	_ = cmd.Flags().StringSliceP("verb", "v", []string{},
		"verbs that apply to the resources contained in the rule",
	)
	_ = cmd.Flags().StringSliceP("resource", "r", []string{},
		"resources that the rule applies to",
	)
	_ = cmd.Flags().StringSliceP("resource-name", "n", []string{},
		"optional resource names that the rule applies to",
	)

	return cmd
}
