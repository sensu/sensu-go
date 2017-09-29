package check

import (
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// CreateCommand adds command that allows user to create new checks
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create NAME",
		Short:        "create new checks",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			isInteractive := flags.NFlag() == 0

			opts := newCheckOpts()

			if isInteractive {
				opts.administerQuestionnaire(false)
			} else {
				opts.withFlags(flags)
				if len(args) > 0 {
					opts.Name = args[0]
				}
			}

			if opts.Org == "" {
				opts.Org = cli.Config.Organization()
			}

			if opts.Env == "" {
				opts.Env = cli.Config.Environment()
			}

			// Apply given arguments to check
			check := types.CheckConfig{}
			opts.Copy(&check)

			if err := check.Validate(); err != nil {
				if !isInteractive {
					cmd.SilenceUsage = false
				}
				return err
			}

			// Ensure that the client is configured to create the check within the
			// corrent context.
			cli.Config.SetOrganization(check.Organization)
			cli.Config.SetEnvironment(check.Environment)

			//
			// TODO:
			//
			// Current validation is a bit too laissez faire. For usability we should
			// determine whether there are assets / handlers / mutators associated w/
			// the check and warn the user if they do not exist yet.

			if err := cli.Client.CreateCheck(&check); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	cmd.Flags().StringP("command", "c", "", "the command the check should run")
	cmd.Flags().StringP("interval", "i", intervalDefault, "interval, in second, at which the check is run")
	cmd.Flags().StringP("subscriptions", "s", "", "comma separated list of topics check requests will be sent to")
	cmd.Flags().String("handlers", "", "comma separated list of handlers to invoke when check fails")
	cmd.Flags().StringP("runtime-assets", "r", "", "comma separated list of assets this check depends on")

	// Mark flags are required for bash-completions
	cmd.MarkFlagRequired("command")
	cmd.MarkFlagRequired("interval")
	cmd.MarkFlagRequired("subscriptions")

	return cmd
}
