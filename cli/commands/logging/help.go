package logging

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "loglevel",
		Short: "change/view loglevel module wise runtime",
		RunE:  helpers.DefaultSubCommandRunE,
	}

	// Add sub-commands
	cmd.AddCommand(
		SetLogLevelForModule(cli),
		SetLogLevelForAllModules(cli),
		ListLogLevels(cli),
		GetModuleLogLevel(cli),
	)

	return cmd
}
