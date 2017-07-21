package logout

import (
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
			// Logout from the configured Sensu instance
			tokens := cli.Config.Tokens()
			if err := cli.Client.Logout(tokens.Refresh); err != nil {
				return err
			}

			// Remove the configured tokens from the local configuration file
			if err := cli.Config.SaveTokens(&types.Tokens{}); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "You have been logout")
			return nil
		},
	}
}
