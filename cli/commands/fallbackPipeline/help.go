package fallbackPipeline

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// HelpCommand defines new pipeline command
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fallback-pipeline",
		Short: "Manage fallback-pipelines",
		RunE:  helpers.DefaultSubCommandRunE,
	}

	// Add sub-commands
	cmd.AddCommand(ListCommand(cli))
	cmd.AddCommand(InfoCommand(cli))
	cmd.AddCommand(DeleteCommand(cli))

	return cmd
}
