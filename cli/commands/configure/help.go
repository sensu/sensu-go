package configure

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure Sensu CLI options",
	}

	// Add sub-commands
	cmd.AddCommand(
		SetOrgCommand(cli),
	)

	return cmd
}
