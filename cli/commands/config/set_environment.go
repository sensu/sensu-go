package config

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/spf13/cobra"
)

// SetEnvCommand given argument changes environment for active profile
func SetEnvCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "set-environment [ENVIRONMENT]",
		Short:        "Set environment for active profile",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			newEnv := args[0]
			if err := cli.Config.SaveEnvironment(newEnv); err != nil {
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
