package command

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// HelpCommand defines new "command" command
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "command",
		Short: "Manage sensuctl commands",
	}

	// Add sub-commands
	cmd.AddCommand(ExecCommand(cli))
	cmd.AddCommand(InstallCommand(cli))

	return cmd
}
