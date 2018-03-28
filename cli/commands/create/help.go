package create

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create sensu resources",
	}

	// Add sub-commands
	cmd.AddCommand(
		CreateCommand(cli),
	)
	return cmd
}
