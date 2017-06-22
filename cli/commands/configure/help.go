package configure

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configure values",
	}

	// Add sub-commands
	cmd.AddCommand(
		SetOrgCommand(cli),
	)

	return cmd
}
