package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/spf13/cobra"
)

// SetTimeoutCommand given argument changes timeout for active profile
func SetTimeoutCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "set-timeout [TIMEOUT]",
		Short:        "Set timeout for active profile in duration format (ex: 15s)",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			newTimeout := args[0]
			newTimeoutDuration, err := time.ParseDuration(newTimeout)
			if err != nil {
				fmt.Fprintf(
					cmd.OutOrStderr(),
					"Unable to parse new timeout with error: %s\n",
					err,
				)
				return err
			}
			if err := cli.Config.SaveTimeout(newTimeoutDuration); err != nil {
				fmt.Fprintf(
					cmd.OutOrStderr(),
					"Unable to write new configuration file with error: %s\n",
					err,
				)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Updated")
			return nil
		},
		Annotations: map[string]string{
			// We want to be able to run this command regardless of whether the CLI
			// has been configured.
			hooks.ConfigurationRequirement: hooks.ConfigurationNotRequired,
		},
	}
}
