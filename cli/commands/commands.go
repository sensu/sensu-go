package commands

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/check"
	"github.com/sensu/sensu-go/cli/commands/configure"
	"github.com/sensu/sensu-go/cli/commands/entity"
	"github.com/sensu/sensu-go/cli/commands/event"
	"github.com/sensu/sensu-go/cli/commands/handler"
	"github.com/spf13/cobra"
)

// AddCommands adds management commands to given command
func AddCommands(rootCmd *cobra.Command, cli *cli.SensuCli) {
	rootCmd.AddCommand(
		configure.Command(cli),

		// Management Commands
		event.HelpCommand(cli),
		check.HelpCommand(cli),
		handler.HelpCommand(cli),
		entity.HelpCommand(cli),
	)

	for _, cmd := range rootCmd.Commands() {
		rootCmd.ValidArgs = append(rootCmd.ValidArgs, cmd.Use)
	}
}
