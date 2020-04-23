package logout

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// Command defines new configuration command
func Command(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "logout",
		Short:        "Logout from sensuctl",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}
			// Logout from the configured Sensu instance
			tokens := cli.Config.Tokens()
			if err := cli.Client.Logout(tokens.Refresh); err != nil {
				return err
			}

			// Remove the configured tokens from the local configuration file
			if err := cli.Config.SaveTokens(&types.Tokens{}); err != nil {
				return err
			}

			// Reset to the default configuration values
			if err := cli.Config.SaveInsecureSkipTLSVerify(false); err != nil {
				return fmt.Errorf(
					"unable to write new configuration file with error: %s",
					err,
				)
			}

			if err := cli.Config.SaveTrustedCAFile(""); err != nil {
				fmt.Fprintln(cmd.OutOrStderr())
				return fmt.Errorf(
					"unable to write new configuration file with error: %s",
					err,
				)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "You have been logged out")
			return nil
		},
	}
}
