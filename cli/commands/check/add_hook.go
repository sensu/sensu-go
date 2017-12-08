package check

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli"
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

// AddCheckHookCommand defines new command to add hooks to a check
func AddCheckHookCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "add-hook CHECKNAME",
		Short:        "add-hook to check",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			isInteractive := flags.NFlag() == 0

			opts := &checkHookOpts{}
			opts.withFlags(flags)

			if len(args) > 0 {
				opts.Check = args[0]
			}
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

	// Mark flags are required for bash-completions
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("hooks")

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
