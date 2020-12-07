package command

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// HelpCommand defines new "command" command
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "command",
		Short: "Manage sensuctl commands",
		RunE:  helpers.DefaultSubCommandRunE,
	}

	// Add sub-commands
	cmd.AddCommand(ExecCommand(cli))
	cmd.AddCommand(InstallCommand(cli))
	cmd.AddCommand(ListCommand(cli))
	cmd.AddCommand(DeleteCommand(cli))

	return cmd
}
