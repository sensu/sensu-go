package tessen

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tessen",
		Short: "Manage tessen configuration",
		RunE:  helpers.DefaultSubCommandRunE,
	}

	// Add sub-commands
	cmd.AddCommand(
		OptInCommand(cli),
		OptOutCommand(cli),
		InfoCommand(cli),
	)

	return cmd
}
