package logging

import (
	"errors"
	"fmt"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/util/logging"
	"github.com/spf13/cobra"
)

// SetLogLevelForModule sets log level for an module
func SetLogLevelForModule(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "set",
		Short:        "set loglevel for an module",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)
			if !isInteractive {
				// Mark flags are required for bash-completions
				_ = cmd.MarkFlagRequired("reason")
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)

			opts := newLoggingOpts()

			if isInteractive {
				if err := opts.administerQuestionnaire(); err != nil {
					return err
				}
			} else {
				opts.withFlags(cmd.Flags())
				if opts.ModuleName == "" && opts.LogLevel == "" {
					return fmt.Errorf("must specify --module or --level")
				}
			}
			var logLevel logging.LogLevelRequest
			if err := opts.Apply(&logLevel); err != nil {
				return err
			}

			if err := cli.Client.SetLogLevel(logLevel); err != nil {
				return err
			}

			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Log level changed")
			return err
		},
	}
	_ = cmd.Flags().StringP("module", "m", "", "module name")
	_ = cmd.Flags().StringP("level", "l", "", "log level")

	helpers.AddInteractiveFlag(cmd.Flags())
	return cmd
}
