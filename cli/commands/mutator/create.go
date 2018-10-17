package mutator

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// CreateCommand adds command that allows the user to create new mutators
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create [NAME]",
		Short:        "create new mutators",
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

			opts := newMutatorOpts()

			if len(args) > 0 {
				opts.Name = args[0]
			}

			opts.Namespace = cli.Config.Namespace()
			if isInteractive {
				if err := opts.administerQuestionnaire(false); err != nil {
					return err
				}
			} else {
				opts.withFlags(cmd.Flags())
			}

			mutator := types.Mutator{}
			opts.Copy(&mutator)

			if err := mutator.Validate(); err != nil {
				if !isInteractive {
					return errors.New("invalid argument(s) received")
				}
				return err
			}

			err := cli.Client.CreateMutator(&mutator)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	cmd.Flags().StringP("command", "c", "", "command to be executed. The event data is passed to the process via STDIN")
	cmd.Flags().String("env-vars", "", "comma separated list of key=value environment variables for the mutator command")
	cmd.Flags().StringP("timeout", "t", "", "execution duration timeout in seconds (hard stop)")
	helpers.AddInteractiveFlag(cmd.Flags())
	return cmd
}
