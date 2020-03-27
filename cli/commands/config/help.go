package config

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Modify sensuctl configuration",
	}

	// Add sub-commands
	cmd.AddCommand(
		SetFormatCommand(cli),
		SetNamespaceCommand(cli),
		SetTimeoutCommand(cli),
		ViewCommand(cli),
	)

	return cmd
}
