package extension

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extension",
		Short: "Manage extension registry",
	}

	// Add sub-commands
	cmd.AddCommand(
		RegisterCommand(cli),
		DeregisterCommand(cli),
		ListCommand(cli),
	)
	return cmd
}
