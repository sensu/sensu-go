package environment

import (
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// CreateCommand adds command that allows users to create new environments
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create NAME",
		Short:        "create new environment",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			isInteractive := flags.NFlag() == 0
			opts := envOpts{}

			if isInteractive {
				opts.administerQuestionnaire(false)
			} else {
				opts.withFlags(flags)
				if len(args) > 0 {
					opts.Name = args[0]
				}
			}

			env := types.Environment{}
			opts.Copy(&env)

			if opts.Org == "" {
				opts.Org = cli.Config.Organization()
			}

			if err := env.Validate(); err != nil {
				if !isInteractive {
					cmd.SilenceUsage = false
				}
				return err
			}

			if opts.Org == "" {
				return fmt.Errorf("an organization must be provided")
			}

			if err := cli.Client.CreateEnvironment(opts.Org, &env); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Created")
			return nil
		},
		Annotations: map[string]string{
			// We want to be able to run this command regardless of whether the CLI
			// has been configured.
			hooks.ConfigurationRequirement: hooks.ConfigurationNotRequired,
		},
	}

	cmd.Flags().StringP("description", "", "", "Description of environment")
	cmd.Flags().StringP("name", "", "", "Name of environment")
	// TODO (Simon): We should be able to use --organization instead but
	// the environment middleware verifies that the env exists in the given org,
	// even if we are actually create this env
	cmd.Flags().StringP("org", "", "", "Name of organization")

	// Mark flags are required for bash-completions
	cmd.MarkFlagRequired("name")

	return cmd
}
