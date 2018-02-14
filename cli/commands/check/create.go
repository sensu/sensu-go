package check

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// CreateCommand adds command that allows user to create new checks
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create [NAME]",
		Short:        "create new checks",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)
			if !isInteractive {
				// Mark flags are required for bash-completions
				_ = cmd.MarkFlagRequired("command")
				_ = cmd.MarkFlagRequired("subscriptions")
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)
			opts := newCheckOpts()

			if len(args) > 0 {
				opts.Name = args[0]
			}

			opts.Org = cli.Config.Organization()
			opts.Env = cli.Config.Environment()

			if isInteractive {
				if err := opts.administerQuestionnaire(false); err != nil {
					return err
				}
			} else {
				opts.withFlags(cmd.Flags())
				if opts.Interval != "" && opts.Cron != "" {
					return fmt.Errorf("cannot specify --interval and --cron at the same time")
				}
				if opts.Interval == "" && opts.Cron == "" {
					return fmt.Errorf("must specify --interval or --cron")
				}
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
	cmd.Flags().String("cron", "", "the cron schedule at which the check is run")
	cmd.Flags().String("handlers", "", "comma separated list of handlers to invoke when check fails")
	cmd.Flags().StringP("interval", "i", "", "interval, in seconds, at which the check is run")
	cmd.Flags().StringP("runtime-assets", "r", "", "comma separated list of assets this check depends on")
	cmd.Flags().String("proxy-entity-id", "", "the check proxy entity, used to create a proxy entity for an external resource")
	cmd.Flags().BoolP("publish", "p", true, "publish check requests")
	cmd.Flags().BoolP("stdin", "", false, "accept event data via STDIN")
	cmd.Flags().StringP("subscriptions", "s", "", "comma separated list of topics check requests will be sent to")
	cmd.Flags().StringP("timeout", "t", "", "timeout, in seconds, at which the check has to run")
	cmd.Flags().String("ttl", "", "time to live in seconds for which a check result is valid")
	cmd.Flags().String("high-flap-threshold", "", "flap detection high threshold (percent state change) for the check")
	cmd.Flags().String("low-flap-threshold", "", "flap detection low threshold (percent state change) for the check")

	helpers.AddInteractiveFlag(cmd.Flags())
	return cmd
}
