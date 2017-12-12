package environment

import (
	"fmt"

	"github.com/sensu/sensu-go/cli"
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

			if len(args) > 0 {
				opts.Name = args[0]
			}

			opts.Org = cli.Config.Organization()

			if isInteractive {
				if err := opts.administerQuestionnaire(false); err != nil {
					return err
				}
			} else {
				opts.withFlags(flags)
			}

			if opts.Org == "" {
				return fmt.Errorf("an organization must be provided")
			}

			env := types.Environment{}
			opts.Copy(&env)

			if err := env.Validate(); err != nil {
				if !isInteractive {
					cmd.SilenceUsage = false
				}
				return err
			}

			if err := cli.Client.CreateEnvironment(opts.Org, &env); err != nil {
				return err
			}

			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Created")
			return err
		},
	}

	_ = cmd.Flags().StringP("description", "", "", "Description of environment")
	// TODO (Simon): We should be able to use --organization instead but
	// the environment middleware verifies that the env exists in the given org,
	// even if we are actually create this env
	_ = cmd.Flags().StringP("org", "", "", "Name of organization")

	// Mark flags are required for bash-completions
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
