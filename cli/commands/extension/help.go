package extension

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extension",
		Short: "Manage extension registry",
		RunE:  helpers.DefaultSubCommandRunE,
	}

	// Add sub-commands
	cmd.AddCommand(
		RegisterCommand(cli),
		DeregisterCommand(cli),
		ListCommand(cli),
	)
	return cmd
}
