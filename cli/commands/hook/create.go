package hook

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// CreateCommand adds command that allows user to create new hooks
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create [NAME]",
		Short:        "create new hooks",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)
			if !isInteractive {
				// Mark flags are required for bash-completions
				_ = cmd.MarkFlagRequired("command")
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)
			opts := newHookOpts()

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
			}

			// Apply given arguments to hook
			hook := types.HookConfig{}
			opts.Copy(&hook)

			if err := hook.Validate(); err != nil {
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
			// the hook and warn the user if they do not exist yet.

			if err := cli.Client.CreateHook(&hook); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	_ = cmd.Flags().StringP("command", "c", "", "the command the hook should run")
	_ = cmd.Flags().StringP("timeout", "t", timeoutDefault, "timeout, in seconds, at which the hook has to run")
	_ = cmd.Flags().BoolP("stdin", "s", false, "stdin enabled on hook")

	helpers.AddInteractiveFlag(cmd.Flags())
	return cmd
}
