package config

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/spf13/cobra"
)

// SetFormatCommand given argument changes format for active profile
func SetFormatCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "set-format [ENVIRONMENT]",
		Short:        "Set format for active profile",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			newFormat := args[0]
			if err := cli.Config.SaveFormat(newFormat); err != nil {
				fmt.Fprintf(
					cmd.OutOrStderr(),
					"Unable to write new configuration file with error: %s\n",
					err,
				)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
		Annotations: map[string]string{
			// We want to be able to run this command regardless of whether the CLI
			// has been configured.
			hooks.ConfigurationRequirement: hooks.ConfigurationNotRequired,
		},
	}
}
