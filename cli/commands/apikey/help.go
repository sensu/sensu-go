package apikey

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api-key",
		Short: "Manage apikeys",
		RunE:  helpers.DefaultSubCommandRunE,
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
