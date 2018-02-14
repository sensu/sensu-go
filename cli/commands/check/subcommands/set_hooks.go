package subcommands

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

type checkHookOpts struct {
	Check string `survey:"check"`
	Type  string `survey:"type"`
	Hooks string `survey:"hooks"`
}

// SetCheckHooksCommand defines new command to set hooks of a check
func SetCheckHooksCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "set-hooks [CHECKNAME]",
		Short:        "set hooks of a check",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)
			if !isInteractive {
				// Mark flags are required for bash-completions
				_ = cmd.MarkFlagRequired("type")
				_ = cmd.MarkFlagRequired("hooks")
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)

			opts := &checkHookOpts{}
			opts.withFlags(cmd.Flags())
			opts.Check = args[0]

			if isInteractive {
				cmd.SilenceUsage = false
				if err := opts.administerQuestionnaire(); err != nil {
					return err
				}
			}

			if opts.Check == "" {
				return errors.New("must provide name of associated check")
			}

			// Instantiate check hook from input
			checkHook := types.HookList{}
			opts.Copy(&checkHook)

			// Ensure that the given checkHook is valid
			if err := checkHook.Validate(); err != nil {
				return err
			}

			check, err := cli.Client.FetchCheck(opts.Check)
			if err != nil {
				return err
			}

			if err := cli.Client.AddCheckHook(check, &checkHook); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Added")
			return nil
		},
	}

	cmd.Flags().StringP("type", "t", "", "type associated with the hook")
	cmd.Flags().StringP("hooks", "k", "", "comma separated list of hooks associated with the check")

	helpers.AddInteractiveFlag(cmd.Flags())
	return cmd
}

func (opts *checkHookOpts) withFlags(flags *pflag.FlagSet) {
	opts.Type, _ = flags.GetString("type")
	opts.Hooks, _ = flags.GetString("hooks")
}

func (opts *checkHookOpts) administerQuestionnaire() error {
	var qs = []*survey.Question{
		{
			Name: "check",
			Prompt: &survey.Input{
				Message: "Check Name:",
				Default: opts.Check,
			},
			Validate: survey.Required,
		},
		{
			Name: "type",
			Prompt: &survey.Input{
				Message: "Hook Type:",
				Default: opts.Type,
			},
			Validate: survey.Required,
		},
		{
			Name: "hooks",
			Prompt: &survey.Input{
				Message: "Hooks:",
			},
			Validate: survey.Required,
		},
	}

	return survey.Ask(qs, opts)
}

func (opts *checkHookOpts) Copy(checkHook *types.HookList) {
	checkHook.Type = opts.Type
	checkHook.Hooks = helpers.SafeSplitCSV(opts.Hooks)
}
