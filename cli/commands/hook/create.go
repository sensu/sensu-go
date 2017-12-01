package hook

import (
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// CreateCommand adds command that allows user to create new hooks
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create NAME",
		Short:        "create new hooks",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			isInteractive := flags.NFlag() == 0

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
				opts.withFlags(flags)
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

	cmd.Flags().StringP("command", "c", "", "the command the hook should run")
	cmd.Flags().StringP("timeout", "t", timeoutDefault, "timeout, in seconds, at which the hook has to run")
	cmd.Flags().BoolP("stdin", "s", false, "stdin enabled on hook")

	// Mark flags are required for bash-completions
	cmd.MarkFlagRequired("command")

	return cmd
}
