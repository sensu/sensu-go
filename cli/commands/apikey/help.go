package apikey

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api-key",
		Short: "Manage apikeys",
	}

	// Add sub-commands
	cmd.AddCommand(
		GrantCommand(cli),
		RevokeCommand(cli),
		ListCommand(cli),
		InfoCommand(cli),
	)

	return cmd
}
