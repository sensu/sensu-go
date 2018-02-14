package check

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type executionOpts struct {
	Creator       string
	Name          string `survey:"check"`
	Reason        string `survey:"reason"`
	Subscriptions string `survey:"subscriptions"`
}

// ExecuteCommand defines a new command to request a check execution
func ExecuteCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "execute [NAME]",
		Short:        "request a check execution",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)
			if !isInteractive {
				// Mark flags are required for bash-completions
				_ = cmd.MarkFlagRequired("check")
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)
			opts := &executionOpts{}

			if len(args) > 0 {
				opts.Name = args[0]
			}

			if isInteractive {
				cmd.SilenceUsage = false
				if err := opts.administerQuestionnaire(); err != nil {
					return err
				}
			} else {
				opts.withFlags(cmd.Flags())
			}

			if opts.Name == "" {
				return errors.New("must provide name of a check")
			}

			// Instantiate an adhoc request from the input
			adhocRequest := &types.AdhocRequest{}
			opts.Copy(adhocRequest)

			// Add the current user as the creator
			adhocRequest.Creator = helpers.GetCurrentUsername(cli.Config)

			// Ensure that the given checkHook is valid
			if err := adhocRequest.Validate(); err != nil {
				return err
			}

			if err := cli.Client.ExecuteCheck(adhocRequest); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Issued")
			return nil
		},
	}

	cmd.Flags().StringP("check", "c", "", "name of the check")
	cmd.Flags().StringP("reason", "r", "", "optional reason for requesting a check execution")
	cmd.Flags().StringP("subscriptions", "s", "", "optional comma separated list of subscriptions to override the check configuration")

	helpers.AddInteractiveFlag(cmd.Flags())

	return cmd
}

func (opts *executionOpts) withFlags(flags *pflag.FlagSet) {
	if name, _ := flags.GetString("check"); name != "" {
		opts.Name = name
	}
	opts.Reason, _ = flags.GetString("reason")
	opts.Subscriptions, _ = flags.GetString("subscriptions")
}

func (opts *executionOpts) administerQuestionnaire() error {
	var qs = []*survey.Question{
		{
			Name: "check",
			Prompt: &survey.Input{
				Message: "Check Name:",
			},
			Validate: survey.Required,
		},
		{
			Name: "reason",
			Prompt: &survey.Input{
				Message: "Reason:",
				Help:    "Optional reason for requesting a check execution",
			},
		},
		{
			Name: "subscriptions",
			Prompt: &survey.Input{
				Message: "Subscriptions:",
				Help:    "Optional comma separated list of subscriptions to override the check configuration",
			},
		},
	}

	return survey.Ask(qs, opts)
}

func (opts *executionOpts) Copy(req *types.AdhocRequest) {
	req.Name = opts.Name
	req.Reason = opts.Reason
	req.Subscriptions = helpers.SafeSplitCSV(opts.Subscriptions)
}
