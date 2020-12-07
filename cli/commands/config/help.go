package config

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Modify sensuctl configuration",
		RunE:  helpers.DefaultSubCommandRunE,
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
